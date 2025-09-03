package upload

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"meliocool/bytesize/internal/helper"
	"meliocool/bytesize/internal/model/domain"
	"meliocool/bytesize/internal/model/web"
	"meliocool/bytesize/internal/repository"
	"meliocool/bytesize/internal/storage"
	"sync"
)

type UploadServiceImpl struct {
	ChunkRepository     repository.ChunkRepository
	FileRepository      repository.FileRepository
	FileChunkRepository repository.FileChunkRepository
	ChunkStore          storage.ChunkStore
	DB                  *pgxpool.Pool
	Validate            *validator.Validate
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

func (u *UploadServiceImpl) Upload(ctx context.Context, request web.UploadRequest) (web.UploadResponse, error) {
	err := u.Validate.Struct(request)
	if err != nil {
		return web.UploadResponse{}, helper.ErrInvalidInput
	}

	tx, err := u.DB.Begin(ctx)
	file := domain.File{
		Filename:  request.FileName,
		TotalSize: 0,
	}
	createdFile, err := u.FileRepository.Create(ctx, tx, file)
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			return web.UploadResponse{}, helper.ErrInternal
		}
		return web.UploadResponse{}, helper.ErrInternal
	}
	_ = tx.Commit(ctx)

	var totalSize int64 = 0
	var chunksCount int64 = 0
	var uniqueChunksWritten int64 = 0
	var dedupeSavedBytes int64 = 0
	var chunkSize = 4 * 1024 * 1024
	var batchSize = 200
	var workers = 10

	chCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	chunksCh := make(chan chunkItem, 8)
	hashedCh := make(chan hashedChunkItem, 8)
	storedCh := make(chan storedChunkItem, 8)

	var wg sync.WaitGroup

	// * Chunker
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, chunkSize)
		var idxLocal int64 = 0
		for {
			select {
			case <-chCtx.Done():
				return
			default:
			}
			n, readErr := request.Reader.Read(buffer)
			if n > 0 {
				chunkSlice := make([]byte, n)
				copy(chunkSlice, buffer)
				chunk := chunkItem{
					Idx:   idxLocal,
					Bytes: chunkSlice,
					Size:  int64(n),
				}
				select {
				case chunksCh <- chunk:
					idxLocal++
				case <-chCtx.Done():
					return
				}
			}
			if readErr == io.EOF {
				close(chunksCh)
				return
			}
			if readErr != nil {
				select {
				case errCh <- readErr:
				default:
				}
				cancel()
				return
			}
		}
	}()

	// * Hasher
	wg.Add(1)
	go func() {
		defer wg.Done()
		for chunk := range chunksCh {
			chunkBytes := sha256.Sum256(chunk.Bytes)
			chunkStr := hex.EncodeToString(chunkBytes[:])
			hashedChunk := hashedChunkItem{
				Idx:   chunk.Idx,
				Bytes: chunk.Bytes,
				Size:  chunk.Size,
				Hash:  chunkStr,
			}
			select {
			case hashedCh <- hashedChunk:
			case <-chCtx.Done():
				return
			}
		}
		close(hashedCh)
	}()

	var wwg sync.WaitGroup

	// * Store + Upsert Worker Pool
	for i := 0; i < workers; i++ {
		wwg.Add(1)
		go func() {
			defer wwg.Done()
			for chunk := range hashedCh {
				select {
				case <-chCtx.Done():
					return
				default:
				}
				reused := false
				ok, exErr := u.ChunkStore.Exists(chunk.Hash)
				if exErr != nil {
					select {
					case errCh <- exErr:
					default:
					}
					cancel()
					return
				}
				if ok {
					reused = true
				} else {
					chunkReader := bytes.NewReader(chunk.Bytes)
					putErr := u.ChunkStore.Put(chunk.Hash, chunkReader, chunk.Size)
					if putErr != nil {
						select {
						case errCh <- putErr:
						default:
						}
						cancel()
						return
					}
				}
				tx, txErr := u.DB.Begin(chCtx)
				if txErr != nil {
					select {
					case errCh <- txErr:
					default:
					}
					cancel()
					return
				}
				_, _, upsertErr := u.ChunkRepository.Upsert(chCtx, tx, domain.Chunk{Hash: chunk.Hash, Size: chunk.Size})
				if upsertErr != nil {
					_ = tx.Rollback(chCtx)
					select {
					case errCh <- upsertErr:
					default:
					}
					cancel()
					return
				}
				commErr := tx.Commit(chCtx)
				if commErr != nil {
					select {
					case errCh <- commErr:
					default:
					}
					cancel()
					return
				}
				chunkStored := storedChunkItem{
					Idx:    chunk.Idx,
					Hash:   chunk.Hash,
					Size:   chunk.Size,
					Reused: reused,
				}
				select {
				case storedCh <- chunkStored:
				case <-chCtx.Done():
					return
				}
			}
		}()
	}
	go func() {
		wwg.Wait()
		close(storedCh)
	}()

	// * Manifest Batcher
	wg.Add(1)
	go func() {
		defer wg.Done()
		var nextExpectedIdx int64 = 0
		items := make(map[int64]storedChunkItem)
		var pendingManifest []domain.FileChunk
		for item := range storedCh {
			items[item.Idx] = item
			for {
				if fetchItem, ok := items[nextExpectedIdx]; !ok {
					break
				} else {
					delete(items, nextExpectedIdx)
					fileChunk := domain.FileChunk{
						FileID:    createdFile.ID,
						Idx:       fetchItem.Idx,
						ChunkHash: fetchItem.Hash,
						Size:      fetchItem.Size,
					}
					pendingManifest = append(pendingManifest, fileChunk)
					totalSize += fetchItem.Size
					chunksCount++
					if fetchItem.Reused == false {
						uniqueChunksWritten++
					} else {
						dedupeSavedBytes += fetchItem.Size
					}
					nextExpectedIdx++
					if len(pendingManifest) == batchSize {
						tx, txManifestErr := u.DB.Begin(chCtx)
						if txManifestErr != nil {
							select {
							case errCh <- txManifestErr:
							default:
							}
							cancel()
							return
						}
						fcrErr := u.FileChunkRepository.AddChunks(chCtx, tx, createdFile.ID, pendingManifest)
						if fcrErr != nil {
							_ = tx.Rollback(chCtx)
							select {
							case errCh <- fcrErr:
							default:
							}
							cancel()
							return
						}
						fcrCommitErr := tx.Commit(chCtx)
						if fcrCommitErr != nil {
							select {
							case errCh <- fcrCommitErr:
							default:
							}
							cancel()
							return
						}
						pendingManifest = pendingManifest[:0]
					}
				}
			}
		}
		if len(pendingManifest) > 0 {
			tx, txManifestErr := u.DB.Begin(chCtx)
			if txManifestErr != nil {
				select {
				case errCh <- txManifestErr:
				default:
				}
				cancel()
				return
			}
			fcrErr := u.FileChunkRepository.AddChunks(chCtx, tx, createdFile.ID, pendingManifest)
			if fcrErr != nil {
				_ = tx.Rollback(chCtx)
				select {
				case errCh <- fcrErr:
				default:
				}
				cancel()
				return
			}
			fcrCommitErr := tx.Commit(chCtx)
			if fcrCommitErr != nil {
				select {
				case errCh <- fcrCommitErr:
				default:
				}
				cancel()
				return
			}
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
	case <-errCh:
		return web.UploadResponse{}, helper.ErrInternal
	}

	cancel()

	return web.UploadResponse{
		FileID:              createdFile.ID,
		TotalSize:           totalSize,
		ChunksCount:         chunksCount,
		UniqueChunksWritten: uniqueChunksWritten,
		DedupeSavedBytes:    dedupeSavedBytes,
	}, nil
}
