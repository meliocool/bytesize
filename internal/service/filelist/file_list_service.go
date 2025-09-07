package filelist

import "context"

type FileListService interface {
	List(ctx context.Context) ([]FileDTO, error)
}
