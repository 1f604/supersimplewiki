package main

import (
	"net/http"
	"sync"
	"time"
)

var pageEditLockMap = make(map[string]*EditLockInfo)

var grab_mut sync.Mutex

type EditLockInfo struct {
	Expires  time.Time
	Username string
}

func canShowEditPage(pageid string, r *http.Request) (string, bool) {
	grab_mut.Lock()
	defer grab_mut.Unlock()

	username := getUsernameFromRequest(r)

	lockInfoPtr, ok := pageEditLockMap[pageid]

	if !ok || lockInfoPtr.Expires.Before(time.Now()) { // it's not currently locked, so lock it
		lockInfo := EditLockInfo{
			Expires:  time.Now().Add(3 * time.Second),
			Username: username,
		}
		pageEditLockMap[pageid] = &lockInfo
		return "NULL", true
	}

	// It is currently locked, so return false
	return lockInfoPtr.Username, false
}

func extendEditLock(pageid string, w http.ResponseWriter, r *http.Request) {
	grab_mut.Lock()
	defer grab_mut.Unlock()

	lockInfoPtr, ok := pageEditLockMap[pageid]
	if !ok { // Nobody has tried to even edit it yet...so just ignore the request
		writeHTTPNoRefreshResponse(w, 400, "Error: Nobody has tried to edit this page yet.")
		return
	}

	username := getUsernameFromRequest(r)
	// if lock has not expired and the requesting user is different, then ignore the request
	if lockInfoPtr.Expires.After(time.Now()) && username != lockInfoPtr.Username {
		writeHTTPNoRefreshResponse(w, 400, "Error: User "+lockInfoPtr.Username+" is already editing this page.")
		return
	}

	// otherwise, extend the lock time with the requesting user
	lockInfo := EditLockInfo{
		Expires:  time.Now().Add(3 * time.Second),
		Username: username,
	}
	pageEditLockMap[pageid] = &lockInfo
	w.WriteHeader(200)
}

func releaseEditLock(pageid string) {
	// user credential checks are done in updateHandler.
	grab_mut.Lock()
	defer grab_mut.Unlock()
	ptr, ok := pageEditLockMap[pageid] // need to check for nil pointer dereference here
	if !ok {                           // If trying to release lock for page that is not locked
		return // just do nothing, since there is no lock to release
	}
	ptr.Expires = time.Time{}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	user, can_edit := canShowEditPage(title, r)
	if !can_edit {
		// page is being edited
		writeHTTPNoRefreshResponse(w, 400, "Error: User "+user+" is currently editing page "+title+".")
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, r, "edit", p)
}
