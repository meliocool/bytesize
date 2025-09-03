package upload

import (
	"context"
	"meliocool/bytesize/internal/model/web"
)

type UploadService interface {
	Upload(ctx context.Context, request web.UploadRequest) (web.UploadResponse, error)
}
