package pagelib

import (
	"net/http"
	"os"
	"time"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

func (p *Page) update() error {
	filename := globals.OS_page_path + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request, title string) {
	filename := globals.OS_page_path + title + ".txt"
	actual_csum, err := util.GetSHA1sumOfFile(filename)
	requestID := r.FormValue("requestID")

	// 1. if page doesn't exist, then don't update it.
	if err != nil {
		util.WriteHTTPNoRefreshResponse(w, 400, "Error: Failed to update page "+title+", because it doesn't exist.")
		return
	}

	// 2. if file checksum has changed, don't update it.
	if r.FormValue("OriginalChecksum") != actual_csum {
		util.WriteHTTPNoRefreshResponse(w, http.StatusInternalServerError, "Error: Page "+title+" changed while you were editing it.")
		return
	}

	// 3. check if the page is currently locked for editing.
	ptr, ok := pageEditLockMap[title]
	if !ok || ptr.Expires.Before(time.Now()) {
		// 3a. if page is not locked, or lock has expired, we can go ahead and edit it.
		goto tryToUpdate
	}
	// 3b. otherwise, if the user does not have edit lock, don't allow the update
	if util.GetUsernameFromRequest(r) != ptr.Username {
		util.WriteHTTPNoRefreshResponse(w, 400, "Error: Failed to update page "+title+", because user "+ptr.Username+" is currently editing it.")
		return
	}

	// Otherwise, since the checksums match, try to update it.
tryToUpdate:
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.update()
	if err != nil {
		util.WriteHTTPNoRefreshResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	releaseEditLock(title)
	util.WriteHTTPResponse(w, 200, "Request "+requestID+" was successfully executed.")
}
