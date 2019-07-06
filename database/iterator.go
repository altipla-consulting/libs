package database

import (
	"database/sql"

	"libs.altipla.consulting/errors"
)

// Iterator helps to loop through rows of a collection retrieving a single model each time.
type Iterator struct {
	rows  *sql.Rows
	props []*Property
}

// Close finishes the iteration. Do not use the iterator after closing it.
func (it *Iterator) Close() {
	it.rows.Close()
}

// Next retrieves the next model of the list. It may or may communicate with the
// database depending on the local cache it has in every loop iteration.
//
// When the iterator reaches the end of the collection it returns ErrDone.
func (it *Iterator) Next(model Model) error {
	if err := it.rows.Err(); err != nil {
		return errors.Trace(err)
	}

	if !it.rows.Next() {
		if err := it.rows.Err(); err != nil {
			return errors.Trace(err)
		}

		it.Close()

		return ErrDone
	}

	props := updateModelProps(it.props, model)
	pointers := make([]interface{}, len(props))
	for i, prop := range props {
		pointers[i] = prop.Pointer
	}
	if err := it.rows.Scan(pointers...); err != nil {
		return errors.Trace(err)
	}

	props = updateModelProps(it.props, model)

	return model.Tracking().AfterGet(props)
}
