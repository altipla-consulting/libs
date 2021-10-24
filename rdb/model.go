package rdb

import (
	"encoding/json"
	"reflect"
	"time"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type Model interface {
	// To be implemented by the model struct. It should return the name of the
	// collection like "Users" or "Sites".
	Collection() string

	// Returns the change vector associated to the model when it was last
	// retrieved from the server. Automatically implemented with rdb.ModelTracking.
	// Deprecated: Use Tracking().ChangeVector() instead.
	ChangeVector() string

	// Returns tracking info like expiration or change vector.
	// Automatically implemented with rdb.ModelTracking.
	Tracking() *ModelTracking

	// Automatically implemented with rdb.ModelTracking. Internal method to initialize
	// new models from scratch.
	load(md api.ModelMetadata)
}

type ModelTracking struct {
	changeVector string
	expires      time.Time
}

func (tracking *ModelTracking) load(md api.ModelMetadata) {
	tracking.changeVector = md.ChangeVector
	tracking.expires = md.Expires
}

func (tracking *ModelTracking) ChangeVector() string {
	return tracking.changeVector
}

func (tracking *ModelTracking) Expires() time.Time {
	return tracking.expires
}

func (tracking *ModelTracking) Expire(t time.Time) {
	tracking.expires = t
}

func (tracking *ModelTracking) NeverExpire() {
	tracking.expires = time.Time{}
}

func (tracking *ModelTracking) Tracking() *ModelTracking {
	return tracking
}

type IndexModel struct {
}

func (model *IndexModel) Collection() string        { return "" }
func (model *IndexModel) ChangeVector() string      { return "" }
func (model *IndexModel) load(md api.ModelMetadata) {}
func (model *IndexModel) Tracking() *ModelTracking  { return nil }

func getModelID(model Model) (string, error) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr {
		return "", errors.Errorf("model should be a pointer to a struct: %T", model)
	}
	if rv.Elem().Kind() != reflect.Struct {
		return "", errors.Errorf("model should be a pointer to a struct: %T", model)
	}

	fv := rv.Elem().FieldByName("ID")
	if _, ok := fv.Interface().(string); !ok {
		return "", errors.Errorf("models should have a string ID field: %T", fv.Interface())
	}

	return fv.Interface().(string), nil
}

func serializeModel(model Model) (map[string]interface{}, error) {
	encoded, err := json.Marshal(model)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var read map[string]interface{}
	if err := json.Unmarshal(encoded, &read); err != nil {
		return nil, errors.Trace(err)
	}

	delete(read, "ID")

	md := map[string]string{
		"@collection": model.Collection(),
	}
	tracking := model.Tracking()
	if !tracking.Expires().IsZero() {
		md["@expires"] = tracking.Expires().In(time.UTC).Format(api.DateTimeFormat)
	}
	read["@metadata"] = md

	return read, nil
}

func setModelID(model interface{}, metadata api.ModelMetadata) error {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Ptr {
		return errors.Errorf("model should be a pointer to a struct: %T", model)
	}
	if rv.Elem().Kind() != reflect.Struct {
		return errors.Errorf("model should be a pointer to a struct: %T", model)
	}

	if _, ok := reflect.TypeOf(model).Elem().FieldByName("ID"); !ok {
		return nil
	}

	fv := rv.Elem().FieldByName("ID")
	if _, ok := fv.Interface().(string); !ok {
		return errors.Errorf("models should have a string ID field: %T", fv.Interface())
	}
	fv.Set(reflect.ValueOf(metadata.ID))

	return nil
}

func checkSingleModel(dest interface{}) error {
	rt := reflect.TypeOf(dest)
	if rt.Kind() != reflect.Ptr {
		return errors.Errorf("dest should be a pointer to a model, to able to initialize nil models: %T", dest)
	}
	if rt.Elem().Kind() != reflect.Ptr {
		return errors.Errorf("dest should be a pointer to a model, which should be a pointer to a struct itself: %T", dest)
	}
	if rt.Elem().Elem().Kind() != reflect.Struct {
		return errors.Errorf("dest should be a pointer to a model, which should be a struct itself: %T", dest)
	}
	return nil
}

func createModel(dest interface{}, result api.Result) (Model, error) {
	rv := reflect.ValueOf(dest)
	if rv.Elem().IsNil() {
		rv.Elem().Set(reflect.New(reflect.TypeOf(dest).Elem().Elem()))
	}

	metadata := api.ModelMetadata{
		ChangeVector: result.Metadata("@change-vector"),
		ID:           result.Metadata("@id"),
	}
	expires := result.Metadata("@expires")
	if expires != "" {
		var err error
		metadata.Expires, err = time.Parse(api.DateTimeFormat, expires)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	if err := setModelID(rv.Elem().Interface(), metadata); err != nil {
		return nil, errors.Trace(err)
	}

	model, ok := rv.Elem().Interface().(Model)
	if ok {
		model.load(metadata)
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if err := json.Unmarshal(encoded, dest); err != nil {
		return nil, errors.Trace(err)
	}

	if ok {
		if result.Metadata("@collection") != "" && model.Collection() != result.Metadata("@collection") {
			return nil, errors.Errorf("expected collection %s, got %s", result.Metadata("@collection"), model.Collection())
		}
	}

	return model, nil
}
