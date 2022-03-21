package main

import (
	"net/http"
	"sync"
	"time"
)

var pageEditLockMap = make(map[string]time.Time)

var grab_mut sync.Mutex

func tryGrabEditLock(pageid string) bool {
	grab_mut.Lock()
	defer grab_mut.Unlock()

	locked_until, ok := pageEditLockMap[pageid]
	if !ok || locked_until.Before(time.Now()) {
		pageEditLockMap[pageid] = time.Now().Add(3 * time.Second)
		return true
	}

	return false
}

func extendEditLock(pageid string) {
	grab_mut.Lock()
	defer grab_mut.Unlock()
	pageEditLockMap[pageid] = time.Now().Add(3 * time.Second)
}

func releaseEditLock(pageid string) {
	grab_mut.Lock()
	defer grab_mut.Unlock()
	pageEditLockMap[pageid] = time.Time{}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	can_show_edit_page := tryGrabEditLock(title)
	if !can_show_edit_page {
		// page is being edited
		w.Write([]byte("Error: Someone is currently editing page " + title + "."))
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, r, "edit", p)
}
