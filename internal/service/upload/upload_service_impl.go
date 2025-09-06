package upload

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/domain"
	"meliocool/bytesize/internal/model/web"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/storage"
	"sync"
	"time"
)

type UploadServiceImpl struct {
	ChunkRepository     repository.ChunkRepository
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	ChunkStore          storage.ChunkStore
	DB                  *pgxpool.Pool
	Validate            *validator.Validate
	Logger              *slog.Logger
}

func NewUploadService(
	chunkRepo repository.ChunkRepository,
	fileRepo repository.FileRepository,
	fileChunkRepo repository.FileChunkRepository,
	chunkStore storage.ChunkStore,
	db *pgxpool.Pool,
	validate *validator.Validate,
	logger *slog.Logger,
) UploadService {
	return &UploadServiceImpl{
		ChunkRepository:     chunkRepo,
		FileRepository:      fileRepo,
		FileChunkRepository: fileChunkRepo,
		ChunkStore:          chunkStore,
		DB:                  db,
		Validate:            validate,
		Logger:              logger,
	}
}

type chunkItem struct {
	Idx   int64
	Bytes []byte
	Size  int64
}

type hashedChunkItem struct {
	Idx   int64
	Bytes []byte
	Size  int64
	Hash  string
}

type storedChunkItem struct {
	Idx    int64
	Hash   string
	Size   int64
	Reused bool
}

type uploadCounters struct {
	TotalSize           int64
	ChunksCount         int64
	UniqueChunksWritten int64
	DedupeSavedBytes    int64
}

// * createFileRow inserts the initial file row
func (u *UploadServiceImpl) createFileRow(ctx context.Context, filename string) (domain.File, error) {
	tx, err := u.DB.Begin(ctx)
	if err != nil {
		return domain.File{}, err
	}
	file := domain.File{Filename: filename, TotalSize: 0}
	createdFile, err := u.FileRepository.Create(ctx, tx, file)
	if err != nil {
		_ = tx.Rollback(ctx)
		return domain.File{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.File{}, err
	}
	return createdFile, nil
}

// * startChunker reads from io.Reader and sends chunks into out channel.
func (u *UploadServiceImpl) startChunker(
	chCtx context.Context,
	wg *sync.WaitGroup,
	r io.Reader,
	out chan<- chunkItem,
	errCh chan<- error,
	chunkSize int,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)

		buffer := make([]byte, chunkSize)
		var idx int64 = 0

		for {
			select {
			case <-chCtx.Done():
				return
			default:
			}

			n, readErr := r.Read(buffer)
			if n > 0 {
				chunkCopy := make([]byte, n)
				copy(chunkCopy, buffer[:n])
				ch := chunkItem{Idx: idx, Bytes: chunkCopy, Size: int64(n)}
				select {
				case out <- ch:
					idx++
				case <-chCtx.Done():
					return
				}
			}

			if readErr == io.EOF {
				return
			}
			if readErr != nil {
				select {
				case errCh <- readErr:
				default:
				}
				return
			}
		}
	}()
}

// startHasher consumes chunkItem, hashes it, then sends hashedChunkItem.
func (u *UploadServiceImpl) startHasher(
	chCtx context.Context,
	wg *sync.WaitGroup,
	in <-chan chunkItem,
	out chan<- hashedChunkItem,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)

		for ch := range in {
			sum := sha256.Sum256(ch.Bytes)
			hash := hex.EncodeToString(sum[:])
			hc := hashedChunkItem{
				Idx:   ch.Idx,
				Bytes: ch.Bytes,
				Size:  ch.Size,
				Hash:  hash,
			}
			select {
			case out <- hc:
			case <-chCtx.Done():
				return
			}
		}
	}()
}

// startStoreWorkers runs a pool of workers that put chunks in FS + Metadata in DB.
func (u *UploadServiceImpl) startStoreWorkers(
	chCtx context.Context,
	wwg *sync.WaitGroup,
	in <-chan hashedChunkItem,
	out chan<- storedChunkItem,
	errCh chan<- error,
	workers int,
) {
	for i := 0; i < workers; i++ {
		wwg.Add(1)
		go func() {
			defer wwg.Done()
			for ch := range in {
				select {
				case <-chCtx.Done():
					return
				default:
				}

				reused := false
				ok, exErr := u.ChunkStore.Exists(ch.Hash)
				if exErr != nil {
					select {
					case errCh <- exErr:
					default:
					}
					return
				}
				if ok {
					reused = true
				} else {
					reader := bytes.NewReader(ch.Bytes)
					if putErr := u.ChunkStore.Put(ch.Hash, reader, ch.Size); putErr != nil {
						select {
						case errCh <- putErr:
						default:
						}
						return
					}
				}

				tx, txErr := u.DB.Begin(chCtx)
				if txErr != nil {
					select {
					case errCh <- txErr:
					default:
					}
					return
				}
				_, _, upsertErr := u.ChunkRepository.Upsert(chCtx, tx, domain.Chunk{Hash: ch.Hash, Size: ch.Size})
				if upsertErr != nil {
					_ = tx.Rollback(chCtx)
					select {
					case errCh <- upsertErr:
					default:
					}
					return
				}
				if err := tx.Commit(chCtx); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}

				sc := storedChunkItem{Idx: ch.Idx, Hash: ch.Hash, Size: ch.Size, Reused: reused}
				select {
				case out <- sc:
				case <-chCtx.Done():
					return
				}
			}
		}()
	}
}

// closes storedCh
func closeStoredWhenWorkersDone(wwg *sync.WaitGroup, out chan<- storedChunkItem) {
	go func() {
		wwg.Wait()
		close(out)
	}()
}

// runManifestBatcher consumes storedChunkItem, batches inserts into DB, and updates totals.
func (u *UploadServiceImpl) runManifestBatcher(
	chCtx context.Context,
	wg *sync.WaitGroup,
	in <-chan storedChunkItem,
	fileID uuid.UUID,
	batchSize int,
	totals *uploadCounters,
	errCh chan<- error,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		var nextIdx int64 = 0
		items := make(map[int64]storedChunkItem)
		var pending []domain.FileChunk

		flush := func() bool {
			tx, txErr := u.DB.Begin(chCtx)
			if txErr != nil {
				select {
				case errCh <- txErr:
				default:
				}
				return false
			}
			err := u.FileChunkRepository.AddChunks(chCtx, tx, fileID, pending)
			if err != nil {
				_ = tx.Rollback(chCtx)
				select {
				case errCh <- err:
				default:
				}
				return false
			}
			if err := tx.Commit(chCtx); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return false
			}
			pending = pending[:0]
			return true
		}

		for item := range in {
			items[item.Idx] = item
			for {
				fetch, ok := items[nextIdx]
				if !ok {
					break
				}
				delete(items, nextIdx)
				fc := domain.FileChunk{
					FileID:    fileID,
					Idx:       fetch.Idx,
					ChunkHash: fetch.Hash,
					Size:      fetch.Size,
				}
				pending = append(pending, fc)

				// update counters
				totals.TotalSize += fetch.Size
				totals.ChunksCount++
				if !fetch.Reused {
					totals.UniqueChunksWritten++
				} else {
					totals.DedupeSavedBytes += fetch.Size
				}

				nextIdx++
				if len(pending) == batchSize {
					if !flush() {
						return
					}
				}
			}
		}

		if len(pending) > 0 {
			_ = flush()
		}
	}()
}

// waits for all goroutines or an error.
func waitForPipeline(wg *sync.WaitGroup, errCh <-chan error) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case err := <-errCh:
		return err
	}
}

// updates the file row with final total size.
func (u *UploadServiceImpl) updateFileTotals(ctx context.Context, fileID uuid.UUID, totalSize int64) error {
	tx, err := u.DB.Begin(ctx)
	if err != nil {
		return err
	}
	err = u.FileRepository.UpdateTotals(ctx, tx, fileID, totalSize)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (u *UploadServiceImpl) Upload(ctx context.Context, req web.UploadRequest) (web.UploadResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		return web.UploadResponse{}, helper.ErrInvalidInput
	}

	start := time.Now()
	u.Logger.Info("upload_start", slog.String("filename", req.FileName))

	createdFile, err := u.createFileRow(ctx, req.FileName)
	if err != nil {
		u.Logger.Error("upload_err", slog.String("stage", "create_file_row"), slog.String("filename", req.FileName), slog.Any("err", err))
		return web.UploadResponse{}, helper.ErrInternal
	}

	chCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errCh := make(chan error, 1)
	chunksCh := make(chan chunkItem, 8)
	hashedCh := make(chan hashedChunkItem, 8)
	storedCh := make(chan storedChunkItem, 8)
	var wg sync.WaitGroup
	var wwg sync.WaitGroup
	totals := &uploadCounters{}

	u.startChunker(chCtx, &wg, req.Reader, chunksCh, errCh, helper.ChunkSize)
	u.startHasher(chCtx, &wg, chunksCh, hashedCh)
	u.startStoreWorkers(chCtx, &wwg, hashedCh, storedCh, errCh, helper.Workers)
	closeStoredWhenWorkersDone(&wwg, storedCh)
	u.runManifestBatcher(chCtx, &wg, storedCh, createdFile.ID, helper.BatchSize, totals, errCh)

	if err := waitForPipeline(&wg, errCh); err != nil {
		u.Logger.Error("upload_err", slog.String("stage", "pipeline"), slog.String("file_id", createdFile.ID.String()), slog.Any("err", err))
		return web.UploadResponse{}, helper.ErrInternal
	}

	if err := u.updateFileTotals(ctx, createdFile.ID, totals.TotalSize); err != nil {
		u.Logger.Error("upload_err", slog.String("stage", "update_totals"), slog.String("file_id", createdFile.ID.String()), slog.Any("err", err))
		return web.UploadResponse{}, helper.ErrInternal
	}

	u.Logger.Info(
		"upload_ok",
		slog.String("file_id", createdFile.ID.String()),
		slog.Int64("total_size", totals.TotalSize),
		slog.Int64("chunks_count", totals.ChunksCount),
		slog.Int64("unique_chunks_written", totals.UniqueChunksWritten),
		slog.Int64("dedupe_saved_bytes", totals.DedupeSavedBytes),
		slog.Duration("took", time.Since(start)), // ➐
	)

	return web.UploadResponse{
		FileID:              createdFile.ID,
		TotalSize:           totals.TotalSize,
		ChunksCount:         totals.ChunksCount,
		UniqueChunksWritten: totals.UniqueChunksWritten,
		DedupeSavedBytes:    totals.DedupeSavedBytes,
	}, nil
}
