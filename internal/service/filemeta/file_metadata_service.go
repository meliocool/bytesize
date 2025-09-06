package filemeta

import (
	"context"
	"github.com/google/uuid"
)

type FileMetaDataService interface {
	GetMeta(ctx context.Context, fileID uuid.UUID) (MetaDataDTO, error)
}
