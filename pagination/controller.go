package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/speps/go-hashids"

	"libs.altipla.consulting/database"
	"libs.altipla.consulting/errors"
	"libs.altipla.consulting/rdb"
)

const (
	// DefaultMaxPageSize is the maximum number of items returned in a page by default.
	DefaultMaxPageSize = 1000

	hashAlphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
)

var (
	h *hashids.HashID
)

func init() {
	hd := hashids.NewData()
	hd.Alphabet = hashAlphabet
	hd.Salt = "libs.altipla.consulting/pagination"
	var err error
	h, err = hashids.NewWithData(hd)
	if err != nil {
		log.Fatal(err)
	}
}

// Controller is implemented by all the pagination controllers of this package.
type Controller interface {
	setMaxPageSize(maxPageSize int32)
	rdbStorage() *rdbStorage
	setPageSize(pageSize int32)
	setToken(token string)
	setPage(page int32, checksum uint32)
}

// ControllerOption configures a paginator.
type ControllerOption func(Controller)

// WithMaxPageSize changes the default maximum page size of 1000.
func WithMaxPageSize(maxPageSize int32) ControllerOption {
	if maxPageSize < 1 {
		panic("cannot configure a max page size less than 1")
	}

	return func(ctrl Controller) {
		ctrl.setMaxPageSize(maxPageSize)
	}
}

// WithRDBInclude brings linked models when querying the database. Cannot be
// used with NewSQL.
func WithRDBInclude(includes ...rdb.IncludeOption) InputAdapter {
	return func(ctrl Controller) {
		storage := ctrl.rdbStorage()
		if storage == nil {
			panic("cannot use WithRDBInclude in a SQL paginator")
		}

		storage.includes = append(storage.includes, includes...)
	}
}

// NewRDBToken creates a paginator for a RavenDB query using tokens.
func NewRDBToken(q *rdb.Query, input InputAdapter, opts ...ControllerOption) *TokenController {
	ctrl := &TokenController{
		sharedController: &sharedController{
			storage:     newRDBStorage(q),
			maxPageSize: DefaultMaxPageSize,
		},
	}
	for _, opt := range opts {
		opt(ctrl)
	}
	input(ctrl)
	return ctrl
}

// NewRDBToken creates a paginator for a RavenDB query using page numbers.
func NewRDBPaged(q *rdb.Query, input InputAdapter, opts ...ControllerOption) *PagedController {
	ctrl := &PagedController{
		sharedController: &sharedController{
			storage:     newRDBStorage(q),
			maxPageSize: DefaultMaxPageSize,
		},
	}
	for _, opt := range opts {
		opt(ctrl)
	}
	input(ctrl)
	return ctrl
}

// NewSQLToken creates a paginator for a MySQL query using tokens.
func NewSQLToken(q *database.Collection, input InputAdapter, opts ...ControllerOption) *TokenController {
	ctrl := &TokenController{
		sharedController: &sharedController{
			storage:     newSQLStorage(q),
			maxPageSize: DefaultMaxPageSize,
		},
	}
	for _, opt := range opts {
		opt(ctrl)
	}
	input(ctrl)
	return ctrl
}

// NewSQLToken creates a paginator for a MySQL query using page numbers.
func NewSQLPaged(q *database.Collection, input InputAdapter, opts ...ControllerOption) *PagedController {
	ctrl := &PagedController{
		sharedController: &sharedController{
			storage:     newSQLStorage(q),
			maxPageSize: DefaultMaxPageSize,
		},
	}
	for _, opt := range opts {
		opt(ctrl)
	}
	input(ctrl)
	return ctrl
}

type sharedController struct {
	storage          storageAdapter
	maxPageSize      int32
	pageSize         int32
	checksum         uint32
	start, totalSize int64
	fetchSize        int64
}

func (ctrl *sharedController) setMaxPageSize(maxPageSize int32) {
	ctrl.maxPageSize = maxPageSize
}

func (ctrl *sharedController) rdbStorage() *rdbStorage {
	storage, _ := ctrl.storage.(*rdbStorage)
	return storage
}

func (ctrl *sharedController) setPageSize(pageSize int32) {
	if pageSize < 0 {
		pageSize = 0
	}
	if pageSize == 0 {
		pageSize = 100
	}
	if pageSize > ctrl.maxPageSize {
		pageSize = ctrl.maxPageSize
	}
	ctrl.pageSize = pageSize
}

// OutOfBounds returns true if the requested page is out of bounds.
func (ctrl *sharedController) OutOfBounds() bool {
	return ctrl.start > 0 && ctrl.start >= ctrl.totalSize
}

// HasNextPage returns true if there is a next page.
func (ctrl *sharedController) HasNextPage() bool {
	end := ctrl.start + ctrl.fetchSize
	return ctrl.totalSize > end
}

// HasPrevPage returns true if there is a previous page.
func (ctrl *sharedController) HasPrevPage() bool {
	return ctrl.start > 0
}

// PageSize returns the page size.
func (ctrl *sharedController) PageSize() int32 {
	return ctrl.pageSize
}

// TotalSize returns the total results of the query.
func (ctrl *sharedController) TotalSize() int64 {
	return ctrl.totalSize
}

// Checksum returns the internal checksum that must validate to perform the query.
func (ctrl *sharedController) Checksum() uint32 {
	return ctrl.checksum
}

type TokenController struct {
	*sharedController
	token string
}

func (ctrl *TokenController) setToken(token string) {
	ctrl.token = token
}

func (ctrl *TokenController) setPage(page int32, checksum uint32) {
}

// Fetch obtains the requested page of items.
func (ctrl *TokenController) Fetch(ctx context.Context, models interface{}) error {
	// Checksum the query including the page size.
	ctrl.checksum = ctrl.storage.checksum(ctrl.pageSize)

	if ctrl.token != "" {
		decoded, err := h.DecodeInt64WithError(ctrl.token)
		if err != nil {
			return errors.Wrapf(ErrInvalidToken, "cannot decode token: %v: %v", ctrl.token, err)
		}
		if len(decoded) != 2 {
			return errors.Wrapf(ErrInvalidToken, "invalid number of parts inside the token: %v: %v", ctrl.token, len(decoded))
		}

		ctrl.start = decoded[1]

		if int64(ctrl.checksum) != decoded[0] {
			return errors.Wrapf(ErrChecksumMismatch, "checksum mismatch: %v: %v, expected %v", ctrl.token, decoded[0], ctrl.checksum)
		}
	}
	if ctrl.start < 0 {
		ctrl.start = 0
	}

	var err error
	ctrl.totalSize, err = ctrl.storage.fetch(ctx, models, ctrl.start, ctrl.pageSize)
	if err != nil {
		return errors.Trace(err)
	}
	ctrl.fetchSize = int64(reflect.ValueOf(models).Elem().Len())

	return nil
}

// NextPageToken returns a token that can be used to fetch the next page.
func (ctrl *TokenController) NextPageToken() string {
	end := ctrl.start + ctrl.fetchSize
	if ctrl.totalSize > end {
		token, err := h.EncodeInt64([]int64{int64(ctrl.checksum), end})
		if err != nil {
			panic(err)
		}
		return token
	}

	return ""
}

// PrevPageToken returns a token that can be used to fetch the previous page.
func (ctrl *TokenController) PrevPageToken() string {
	prev := ctrl.start - int64(ctrl.pageSize)
	if ctrl.start > 0 {
		token, err := h.EncodeInt64([]int64{int64(ctrl.checksum), prev})
		if err != nil {
			panic(err)
		}
		return token
	}

	return ""
}

// NextPageURL modifies the URL to point to the next page.
func (ctrl *TokenController) NextPageURL(u *url.URL) *url.URL {
	if !ctrl.HasNextPage() {
		return nil
	}

	qs := u.Query()
	qs.Set("token", ctrl.NextPageToken())
	u.RawQuery = qs.Encode()

	return u
}

// PrevPageURL modifies the URL to point to the previous page.
func (ctrl *TokenController) PrevPageURL(u *url.URL) *url.URL {
	if !ctrl.HasPrevPage() {
		return nil
	}

	qs := u.Query()
	qs.Set("token", ctrl.PrevPageToken())
	u.RawQuery = qs.Encode()

	return u
}

// NextPageURLString returns a new URL based on the current one for the next page.
func (ctrl *TokenController) NextPageURLString(r *http.Request) string {
	u := new(url.URL)
	*u = *r.URL
	if next := ctrl.NextPageURL(u); next != nil {
		return next.String()
	}
	return ""
}

// PrevPageURLString returns a new URL based on the current one for the previous page.
func (ctrl *TokenController) PrevPageURLString(r *http.Request) string {
	u := new(url.URL)
	*u = *r.URL
	if prev := ctrl.PrevPageURL(u); prev != nil {
		return prev.String()
	}
	return ""
}

type PagedController struct {
	*sharedController
	page int32
}

func (ctrl *PagedController) setToken(token string) {
}

func (ctrl *PagedController) setPage(page int32, checksum uint32) {
	ctrl.page = page
	ctrl.checksum = checksum
}

// Fetch obtains the requested page of items.
func (ctrl *PagedController) Fetch(ctx context.Context, models interface{}) error {
	// Checksum the query including the page size.
	// It is safe to convert the uint32 to a int64 and we will
	// never get a negative number doing so.
	checksum := ctrl.storage.checksum(ctrl.pageSize)
	if ctrl.page > 1 && checksum != ctrl.checksum {
		return errors.Wrapf(ErrChecksumMismatch, "checksum mismatch: %v, expected %v", ctrl.checksum, checksum)
	}
	ctrl.checksum = checksum

	ctrl.start = int64((ctrl.page - 1) * ctrl.pageSize)
	if ctrl.start < 0 {
		ctrl.start = 0
	}

	var err error
	ctrl.totalSize, err = ctrl.storage.fetch(ctx, models, ctrl.start, ctrl.pageSize)
	if err != nil {
		return errors.Trace(err)
	}
	ctrl.fetchSize = int64(reflect.ValueOf(models).Elem().Len())

	return nil
}

// NextPageURL modifies the URL to point to the next page.
func (ctrl *PagedController) NextPageURL(u *url.URL) *url.URL {
	if !ctrl.HasNextPage() {
		return nil
	}

	qs := u.Query()
	qs.Set("page", fmt.Sprintf("%v", ctrl.page+1))
	qs.Set("checksum", fmt.Sprintf("%v", ctrl.Checksum()))
	u.RawQuery = qs.Encode()

	return u
}

func (ctrl *PagedController) prevPage() int64 {
	last := ctrl.totalSize / int64(ctrl.pageSize)
	if ctrl.totalSize%int64(ctrl.pageSize) != 0 {
		last++
	}
	prev := int64(ctrl.page) - 1
	if prev > last {
		prev = last
	}
	return prev
}

// PrevPageURL modifies the URL to point to the previous page.
func (ctrl *PagedController) PrevPageURL(u *url.URL) *url.URL {
	if !ctrl.HasPrevPage() {
		return nil
	}

	qs := u.Query()
	if ctrl.page == 2 {
		qs.Del("page")
		qs.Del("checksum")
	} else {
		qs.Set("page", fmt.Sprintf("%v", ctrl.prevPage()))
		qs.Set("checksum", fmt.Sprintf("%v", ctrl.Checksum()))
	}
	u.RawQuery = qs.Encode()

	return u
}

// NextPageURLString returns a new URL based on the current one for the next page.
func (ctrl *PagedController) NextPageURLString(r *http.Request) string {
	u := new(url.URL)
	*u = *r.URL
	if next := ctrl.NextPageURL(u); next != nil {
		return next.String()
	}
	return ""
}

// PrevPageURLString returns a new URL based on the current one for the previous page.
func (ctrl *PagedController) PrevPageURLString(r *http.Request) string {
	u := new(url.URL)
	*u = *r.URL
	if prev := ctrl.PrevPageURL(u); prev != nil {
		return prev.String()
	}
	return ""
}

// StateAsJSONString returns an overview of the paginator state in JSON format
// suitable for use in client components like altipla/ui-v2.
func (ctrl *PagedController) StateAsJSONString() (string, error) {
	repr := struct {
		Next        int64 `json:"next"`
		Prev        int64 `json:"prev"`
		Start       int64 `json:"start"`
		End         int64 `json:"end"`
		TotalSize   int64 `json:"totalSize"`
		OutOfBounds bool  `json:"outOfBounds"`
	}{
		Start:       ctrl.start + 1,
		End:         ctrl.start + ctrl.fetchSize,
		TotalSize:   ctrl.totalSize,
		OutOfBounds: ctrl.OutOfBounds(),
	}
	if ctrl.HasPrevPage() {
		repr.Prev = ctrl.prevPage()
	}
	if ctrl.HasNextPage() {
		repr.Next = int64(ctrl.page + 1)
	}

	b, err := json.Marshal(repr)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(b), nil
}
