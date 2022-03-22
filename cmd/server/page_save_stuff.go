package main

import (
	"net/http"
	"os"
)

func (p *Page) update() error {
	filename := os_page_path + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func updateHandler(w http.ResponseWriter, r *http.Request, title string) {
	// TODO: Add user credential checks here.
	filename := os_page_path + title + ".txt"
	actual_csum, err := getSHA1sumOfFile(filename)

	// 1. if page doesn't exist, then don't update it.
	if err != nil {
		writeHTTPNoRefreshResponse(w, 400, "Error: Failed to update page "+title+", because it doesn't exist.")
		return
	}

	// 2. if file checksum has changed, don't update it.
	if r.FormValue("OriginalChecksum") != actual_csum {
		writeHTTPNoRefreshResponse(w, http.StatusInternalServerError, "Error: Page "+title+" changed while you were editing it.")
		return
	}

	// 3. Otherwise, since the checksums match, try to update it.
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.update()
	if err != nil {
		writeHTTPNoRefreshResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	releaseEditLock(title)
	http.Redirect(w, r, view_path+title, http.StatusFound)
}
