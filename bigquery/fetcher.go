package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

var Done = iterator.Done

type Fetcher interface {
	Fetch(ctx context.Context, query string) (Iterator, error)
}

type Iterator interface {
	Next(dst interface{}) error
	PageInfo() *iterator.PageInfo
}

func NewFetcher(project string) (Fetcher, error) {
	client, err := bigquery.NewClient(context.Background(), project)
	if err != nil {
		return nil, err
	}

	return &productionFetcher{
		client: client,
	}, nil
}

type productionFetcher struct {
	client *bigquery.Client
}

func (u *productionFetcher) Fetch(ctx context.Context, query string) (Iterator, error) {
	it, err := u.client.Query(query).Read(ctx)
	if err != nil {
		return nil, err
	}

	return it, nil
}
