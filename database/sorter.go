package database

// Sorter should be implemented by any generic SQL order we can apply to collections.
type Sorter interface {
	// SQL returns the portion of the code that will be merged to the query.
	SQL() string
}
