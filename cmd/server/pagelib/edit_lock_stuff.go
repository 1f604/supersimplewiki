package pagelib

import (
	"net/http"
	"regexp"
	"sync"
	"time"

	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

var pageEditLockMap = make(map[string]*EditLockInfo)

var grab_mut sync.Mutex

type EditLockInfo struct {
	Expires  time.Time
	Username string
}

func isPageLockedForEditing(pageid string, r *http.Request) (string, bool) {
	grab_mut.Lock()
	defer grab_mut.Unlock()

	username := util.GetUsernameFromRequest(r)

	lockInfoPtr, ok := pageEditLockMap[pageid]

	if !ok || lockInfoPtr.Expires.Before(time.Now()) { // it's not currently locked, so lock it
		lockInfo := EditLockInfo{
			Expires:  time.Now().Add(3 * time.Second),
			Username: username,
		}
		pageEditLockMap[pageid] = &lockInfo
		return "NULL USER", false
	}

	// It is currently locked
	return lockInfoPtr.Username, true
}

func ExtendEditLock(pageid string, w http.ResponseWriter, r *http.Request) {
	grab_mut.Lock()
	defer grab_mut.Unlock()

	lockInfoPtr, ok := pageEditLockMap[pageid]
	if !ok { // Nobody has tried to even edit it yet...so just ignore the request
		util.WriteHTTPNoRefreshResponse(w, 400, "Error: Nobody has tried to edit this page yet.")
		return
	}

	username := util.GetUsernameFromRequest(r)
	// if lock has not expired and the requesting user is different, then ignore the request
	if lockInfoPtr.Expires.After(time.Now()) && username != lockInfoPtr.Username {
		util.WriteHTTPNoRefreshResponse(w, 400, "Error: User "+lockInfoPtr.Username+" is already editing this page.")
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

// a page ID is a string of 6 random alphanumeric characters (uppercase, lowercase, and numbers)
var validEditDebugUpdatePath = regexp.MustCompile("^/(edit|debug|update)/([A-Za-z0-9]{6})$")

func CheckPageExistsWrapper(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := util.MatchRegex(r.URL.Path, validEditDebugUpdatePath)
		if err != nil {
			util.WriteHTTPNoRefreshResponse(w, 406, "Error: invalid page URL: wrong format.")
			return
		}
		if !CheckPageExists(m[2]) {
			util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
			return
		}
		fn(w, r, m[2])
	}
}

func CheckPageLockedWrapper(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	newfn := func(w http.ResponseWriter, r *http.Request, pageID string) {
		// assume URL is valid and page exists.
		// now we check whether page is locked for editing.
		user, locked := isPageLockedForEditing(pageID, r)
		if locked {
			// page is being edited
			util.WriteHTTPNoRefreshResponse(w, 400, "Error: User "+user+" is currently editing page "+pageID+".")
			return
		}
		fn(w, r, pageID)
	}
	return CheckPageExistsWrapper(newfn)
}
