package main

import (
	"net/http"
	"regexp"

	"github.com/1f604/supersimplewiki/cmd/server/pagelib"
	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

var viewPathRegex = regexp.MustCompile("^/view/([A-Za-z0-9]{4})_([0-9]+)_([0-9a-f]{5})$")
var editDebugPathRegex = regexp.MustCompile("^/(edit|debug)/([A-Za-z0-9]{4})$")

func UnifiedPageHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	query := r.URL.Query()
	var params interface{}
	var capturedGroups []string
	var err error

	// 1. Check if it's a View request
	capturedGroups, err = util.MatchRegex(path, viewPathRegex)
	if err == nil { // it's a view request
		params = pagelib.ViewRequest{
			PageID:  capturedGroups[1],
			RevNo:   capturedGroups[2],
			RevHash: capturedGroups[3],
			Query:   &query,
		}
		goto PageFound
	}

	// 2. Check if it's a Edit or Debug request
	capturedGroups, err = util.MatchRegex(path, editDebugPathRegex)
	if err == nil { // it's an edit or debug request
		params = pagelib.EditDebugRequest{
			Req:    capturedGroups[1],
			PageID: capturedGroups[2],
			Query:  &query,
		}
		goto PageFound
	}

	// 3. Check if it's the home page
	if path == "/" {
		util.WriteHTTPNoRefreshResponse(w, 200, "This is the home page.")
		return
	}

	// Otherwise, return an error
	util.WriteHTTPNoRefreshResponse(w, 404, "Error: Unrecognized URL format.")
	return

PageFound:
	pagelib.CommonPageHandler(w, r, &params)
}
