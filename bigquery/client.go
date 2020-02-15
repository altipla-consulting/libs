package bigquery

import (
	"context"

	"cloud.google.com/go/bigquery"
	"libs.altipla.consulting/errors"
)

// Dataset represents a BigQuery dataset name.
type Dataset string

// Table represents a BigQuery table name.
type Table string

// Inserter helps with batch insertion of new rows in a table.
type Inserter = bigquery.Inserter

// StructSaver helps with struct insertion in a table.
type StructSaver = bigquery.StructSaver

// Client holds a connection to the BigQuery service.
type Client struct {
	bq      *bigquery.Client
	dataset Dataset
}

// NewClient opens a new connection to the BigQuery service.
func NewClient(googleProject string, dataset Dataset) (*Client, error) {
	bq, err := bigquery.NewClient(context.Background(), googleProject)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Client{
		bq:      bq,
		dataset: dataset,
	}, nil
}

// Query prepares a new paginator for the query.
func (client *Client) Query(query *Query) *Pager {
	return &Pager{
		query:   query,
		bq:      client.bq,
		dataset: client.dataset,
	}
}

// Inserter returns a helper to insert new rows into the table.
func (client *Client) Inserter(table Table) *Inserter {
	return client.bq.Dataset(string(client.dataset)).Table(string(table)).Inserter()
}
