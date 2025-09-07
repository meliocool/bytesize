package delete

import (
	"context"
	"github.com/google/uuid"
)

type DeleteService interface {
	Delete(ctx context.Context, id uuid.UUID) (Result, error)
}
