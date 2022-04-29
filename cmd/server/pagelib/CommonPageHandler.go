package pagelib

import (
	"log"
	"net/http"

	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

func CommonPageHandler(w http.ResponseWriter, r *http.Request, page_type int, params interface{}) {
	// regardless of the page type, write the "you are logged in as" string

	if !CheckPageExists(pageID) { // TODO: Fix this to account for the different URL format.
		util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
		return
	}

	switch params.(type) {
	case *ViewRequest:
		ViewHandler(w, r, params.(*ViewRequest))
		return
	case *EditDebugRequest:
		if params.(*EditDebugRequest).Req == "edit" {
			EditHandler(w, r, params.(*EditDebugRequest))
		} else if params.(*EditDebugRequest).Req == "debug" {
			DebugHandler(w, r, params.(*EditDebugRequest))
		} else {
			log.Fatal("CommonPageHandler: This should never happen.")
		}
		return
	default:
		log.Fatal("CommonPageHandler: Unknown type. This should never happen.")
	}
}
