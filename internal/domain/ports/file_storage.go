package ports

import "context"

type FileStorage interface {
	UploadBase64Image(ctx context.Context, base64Data, publicID, folder string) (string, error)
}
