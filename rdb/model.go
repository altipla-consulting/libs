package rdb

import (
	"encoding/json"
	"reflect"

	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb/api"
)

type Model interface {
	// To be implemented by the model struct.
	Collection() string

	// Automatically implemented with rdb.ModelTracking.
	ChangeVector() string
	load(changeVector string)
}

type ModelTracking struct {
	changeVector string
}

func (tracking *ModelTracking) load(changeVector string) {
	tracking.changeVector = changeVector
}

func (tracking *ModelTracking) ChangeVector() string {
	return tracking.changeVector
}

type IndexModel struct {
}

func (model *IndexModel) Collection() string       { return "" }
func (model *IndexModel) ChangeVector() string     { return "" }
func (model *IndexModel) load(changeVector string) {}

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

	read["@metadata"] = map[string]string{
		"@collection": model.Collection(),
	}

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
	if err := setModelID(rv.Elem().Interface(), metadata); err != nil {
		return nil, errors.Trace(err)
	}

	model, ok := rv.Elem().Interface().(Model)
	if ok {
		model.load(metadata.ChangeVector)
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
