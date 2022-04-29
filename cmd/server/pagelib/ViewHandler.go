package pagelib

import (
	"net/http"
	"net/url"
)

type ViewRequest struct {
	PageID  string
	RevNo   string
	RevHash string
	Query   *url.Values
}

type EditDebugRequest struct {
	Req    string
	PageID string
	Query  *url.Values
}

func ViewHandler(w http.ResponseWriter, r *http.Request, vreq *ViewRequest) {
	// TODO: Implement this
}
