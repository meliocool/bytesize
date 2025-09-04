package web

import (
	"context"
	"io"
)

type UploadRequest struct {
	Ctx      context.Context
	FileName string    `validate:"required" json:"filename"`
	Reader   io.Reader `validate:"required"`
}
