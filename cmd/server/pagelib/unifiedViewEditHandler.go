package pagelib

import (
	"net/http"
	"regexp"

	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

var validViewPath = regexp.MustCompile("^/(view)/([A-Za-z0-9_]+)$") // TODO: fix this regex.

func UnifiedViewEditHandler(w http.ResponseWriter, r *http.Request, page_type int) {

	m := validViewPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		util.WriteHTTPNoRefreshResponse(w, 406, "Error: invalid page URL: wrong format.")
		return
	}
	pageID := m[2]
	if !CheckPageExists(pageID) { // TODO: Fix this to account for the different URL format.
		util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
		return
	}

}
