package storage

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"libs.altipla.consulting/errors"
)

type Writer interface {
	WriteFile(ctx context.Context, path string, content []byte) error
}

type productionWriter struct {
	bucket *storage.BucketHandle
}

func NewWriter(bucketName string) (Writer, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &productionWriter{
		bucket: client.Bucket(bucketName),
	}, nil
}

func (w *productionWriter) WriteFile(ctx context.Context, path string, content []byte) error {
	var err error
	for i := 0; i < 3; i++ {
		err = w.writeFileSafe(ctx, path, content)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return errors.Trace(err)
}

func (w *productionWriter) writeFileSafe(ctx context.Context, path string, content []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	writer := w.bucket.Object(path).NewWriter(ctx)

	if _, err := fmt.Fprintf(writer, string(content)); err != nil {
		writer.Close()
		return errors.Wrapf(err, "cannot write file: %s", path)
	}

	if err := writer.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file: %s", path)
	}

	return nil
}
