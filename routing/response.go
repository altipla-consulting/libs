package routing

import (
	"encoding/json"
	"net/http"

	"libs.altipla.consulting/errors"
)

func JSON(w http.ResponseWriter, reply interface{}, opts ...ReplyOption) error {
	for _, opt := range opts {
		opt(w)
	}
	return errors.Trace(json.NewEncoder(w).Encode(reply))
}

type ReplyOption func(w http.ResponseWriter)

func WithStatus(statusCode int) ReplyOption {
	return func(w http.ResponseWriter) {
		w.WriteHeader(statusCode)
	}
}
