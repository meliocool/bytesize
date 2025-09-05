package download

import (
	"context"
	"github.com/google/uuid"
	"io"
)

type DownloadService interface {
	Stream(ctx context.Context, fileID uuid.UUID, w io.Writer) error
}
