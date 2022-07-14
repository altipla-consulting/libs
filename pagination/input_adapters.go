package pagination

import (
	"net/http"
	"strconv"
)

// InputAdapter reads multiple kind of sources to configure the pagination.
type InputAdapter func(Controller)

// FromRequest reads the input parameters from a HTTP request.
func FromRequest(r *http.Request) InputAdapter {
	pageSize, _ := strconv.ParseInt(r.FormValue("page-size"), 10, 32)
	if token := r.FormValue("token"); token != "" {
		return FromToken(int32(pageSize), token)
	}
	page, _ := strconv.ParseInt(r.FormValue("page"), 10, 32)
	checksum, _ := strconv.ParseUint(r.FormValue("checksum"), 10, 32)
	return FromPaged(int32(pageSize), int32(page), uint32(checksum))
}

// TokenPaginationRequest is implemented by the generated Protobuf structs if using
// the Google guidelines for API pagination.
type TokenPaginationRequest interface {
	GetPageSize() int32
	GetPageToken() string
}

// FromAPI configures the pagination from a standard gRPC request message.
func FromAPI(msg TokenPaginationRequest) InputAdapter {
	pageSize := msg.GetPageSize()
	if pageSize == 0 {
		pageSize = 100
	}
	return FromToken(pageSize, msg.GetPageToken())
}

// FromToken directly configures the input parameters of a token pagination.
func FromToken(pageSize int32, token string) InputAdapter {
	return func(ctrl Controller) {
		ctrl.setPageSize(pageSize)
		ctrl.setToken(token)
	}
}

// FromPaged directly configures the input parameters of a paged pagination.
func FromPaged(pageSize int32, page int32, checksum uint32) InputAdapter {
	if page < 1 {
		page = 1
	}

	return func(ctrl Controller) {
		ctrl.setPageSize(pageSize)
		ctrl.setPage(page, checksum)
	}
}

// FromEmpty configures the pagination with an empty input. Used mostly in tests
// or other internal helpers.
func FromEmpty() InputAdapter {
	return func(ctrl Controller) {
		ctrl.setPageSize(0)
	}
}
