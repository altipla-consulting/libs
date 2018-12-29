package database

import (
	"database/sql"
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
	modelProps := updatedProps(it.props, model)

	if err := it.rows.Err(); err != nil {
		return err
	}

	if !it.rows.Next() {
		if err := it.rows.Err(); err != nil {
			return err
		}

		it.Close()

		return ErrDone
	}

	ptrs := make([]interface{}, len(modelProps))
	for i, prop := range modelProps {
		ptrs[i] = prop.Pointer
	}
	if err := it.rows.Scan(ptrs...); err != nil {
		return err
	}

	modelProps = updatedProps(it.props, model)

	return model.Tracking().AfterGet(modelProps)
}
