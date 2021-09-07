package cloudrun

import (
	"encoding/json"
	"fmt"
	"net/http"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/pubsub"
	"libs.altipla.consulting/routing"
)

func makePubSubHandler(handler PubSubHandler) routing.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		in := new(pubsub.PushRequest)
		if err := json.NewDecoder(r.Body).Decode(in); err != nil {
			return errors.Trace(err)
		}
		if err := handler(r.Context(), in.Message); err != nil {
			return errors.Trace(err)
		}

		fmt.Fprintln(w, "ok")
		return nil
	}
}
