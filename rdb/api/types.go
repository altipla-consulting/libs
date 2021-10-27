package api

import "time"

type Database struct {
	DatabaseName  string
	DatabaseState string `json:",omitempty"`
	Disabled      bool   `json:",omitempty"`
	DataDirectory string `json:",omitempty"`
	Indexes       map[string]*Index
	Revisions     *Revisions
}

type IndexesRequest struct {
	Indexes []*Index
}

type Index struct {
	Type                     string
	Name                     string
	Maps                     []string
	Reduce                   *string
	Fields                   map[string]*IndexFieldOptions
	OutputReduceToCollection *string
	AdditionalSources        map[string]string
}

func (index *Index) FieldOrCreate(name string) *IndexFieldOptions {
	if field := index.Fields[name]; field != nil {
		return field
	}
	index.Fields[name] = new(IndexFieldOptions)
	return index.Fields[name]
}

type IndexFieldOptions struct {
	Indexing string `json:",omitempty"`
	Storage  string `json:",omitempty"`
}

type Revisions struct {
	Default     *RevisionConfig
	Collections map[string]*RevisionConfig
}

type RevisionConfig struct {
	MinimumRevisionsToKeep   int64
	MinimumRevisionAgeToKeep Duration
	Disabled, PurgeOnDelete  bool
}

type ModelMetadata struct {
	ID           string `json:"Id"`
	ChangeVector string
	Expires      time.Time
}

type Results struct {
	Results  []Result
	Includes map[string]Result

	// Only available in query results.
	TotalResults     int64
	SkippedResults   int64
	DurationInMs     int64
	IndexName        string
	IsStale          bool
	CappedMaxResults *int64
}

type Result map[string]interface{}

func (result Result) Metadata(key string) string {
	m := result["@metadata"].(map[string]interface{})
	value, ok := m[key]
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func (result Result) MetadataBool(key string) bool {
	m := result["@metadata"].(map[string]interface{})
	str, ok := m[key]
	if !ok {
		return false
	}
	return str.(bool)
}

func (result Result) DirectMetadata(key string) string {
	val, _ := result[key].(string)
	return val
}

type BulkCommands struct {
	Commands []BatchCommand
}

type BatchCommand interface {
	isBatchCommand()
}

type PutCommand struct {
	ID           string `json:"Id,omitempty"`
	ChangeVector *string
	Document     interface{}

	// Always set to "PUT"
	Type string
}

func (cmd *PutCommand) isBatchCommand() {}

type DeleteCommand struct {
	ID           string  `json:"Id,omitempty"`
	ChangeVector *string `json:",omitempty"`

	// Always set to "DELETE"
	Type string
}

func (cmd *DeleteCommand) isBatchCommand() {}

type DeletePrefixCommand struct {
	ID string `json:"ID"`

	// Always set to true
	IDPrefixed bool `json:"IdPrefixed"`

	// Always set to "DELETE"
	Type string
}

func (cmd *DeletePrefixCommand) isBatchCommand() {}

type Query struct {
	Query                         string
	QueryParameters               map[string]interface{} `json:",omitempty"`
	WaitForNonStaleResults        bool                   `json:",omitempty"`
	WaitForNonStaleResultsTimeout string                 `json:",omitempty"`
}

type Patch struct {
	Query *Query
}

type Operation struct {
	ID int64 `json:"OperationId"`
}

const (
	OperationStatusInProgress = "InProgress"
	OperationStatusCompleted  = "Completed"
	OperationStatusFaulted    = "Faulted"
)

type OperationStatus struct {
	Status   string
	Progress *OperationProgress
	Result   *OperationResult
}

type OperationProgress struct {
	Processed int64
	Total     int64
}

type OperationResult struct {
	Total int64

	// Only filled if it fails
	Message string
	Error   string
}

type Counters struct {
	Counters []*Counter
}

type Counter struct {
	DocumentID  string `json:"DocumentId"`
	CounterName string
	TotalValue  int64
}

type CounterOperations struct {
	Documents []*CounterOperationDocument
}

type CounterOperationDocument struct {
	DocumentID string `json:"DocumentId"`
	Operations []*CounterOperation
}

type CounterOperationType string

const (
	CounterOperationTypeIncrement = CounterOperationType("Increment")
	CounterOperationTypeDelete    = CounterOperationType("Delete")
	CounterOperationTypeGet       = CounterOperationType("Get")
)

type CounterOperation struct {
	CounterName string
	Delta       int64
	Type        CounterOperationType
}

type DocsRequest struct {
	IDs []string `json:"Ids"`
}
