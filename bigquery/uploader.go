package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
)

type StructSaver = bigquery.StructSaver
type NullString = bigquery.NullString

type Uploader interface {
	Upload(ctx context.Context, dataset, table string, savers []*bigquery.StructSaver) error
}

func NewUploader(project string) (Uploader, error) {
	client, err := bigquery.NewClient(context.Background(), project)
	if err != nil {
		return nil, err
	}

	return &productionUploader{
		client: client,
	}, nil
}

type productionUploader struct {
	client *bigquery.Client
}

func (u *productionUploader) Upload(ctx context.Context, dataset, table string, savers []*bigquery.StructSaver) error {
	return u.client.Dataset(dataset).Table(table).Uploader().Put(ctx, savers)
}
