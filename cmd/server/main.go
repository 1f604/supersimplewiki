// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/1f604/supersimplewiki/cmd/server/editors"
	"github.com/1f604/supersimplewiki/cmd/server/globals"
	"github.com/1f604/supersimplewiki/cmd/server/pagelib"
	util "github.com/1f604/supersimplewiki/cmd/server/util"
)

const (
	login_word   = "login"
	signup_word  = "signup"
	md2html_word = "md2html"
	view_path    = "/view/"
	edit_path    = "/edit/"
	debug_path   = "/debug/"
	update_path  = "/update/"
	lock_path    = "/lock_page/"
	login_path1  = "/" + login_word
	login_path2  = "/" + login_word + "/"
	signup_path1 = "/" + signup_word
	signup_path2 = "/" + signup_word + "/"
	md2html_path = "/" + md2html_word + "/"
)

var validViewPath = regexp.MustCompile("^/(view)/([A-Za-z0-9_]+)$") // TODO: fix this regex.

// the view handler needs its own logic to validate URL page ID due to the unique format
func viewHandler(w http.ResponseWriter, r *http.Request) {
	m := validViewPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		util.WriteHTTPNoRefreshResponse(w, 406, "Error: invalid page URL: wrong format.")
		return
	}
	pageID := m[2]
	if !pagelib.CheckPageExists(pageID) { // TODO: Fix this to account for the different URL format.
		util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
		return
	}

	p, err := pagelib.LoadPage(pageID)
	if err != nil {
		http.Redirect(w, r, edit_path+pageID, http.StatusFound)
		return
	}
	pagelib.RenderTemplate(w, r, "view", p)
}

func lockpageHandler(w http.ResponseWriter, r *http.Request) {
	pageid := r.URL.Path[len(lock_path):]
	pagelib.ExtendEditLock(pageid, w, r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		util.WriteHTTPNoRefreshResponse(w, 404, "404 page not found")
		return
	}
	util.WriteHTTPNoRefreshResponse(w, 200, "This is the home page.")
}

const (
	lNet  = "tcp"
	lAddr = ":12345"
)

func main() {
	if _, err := net.Listen(lNet, lAddr); err != nil {
		fmt.Println("Another instance of supersimplewiki is currently running!")
		return
	}

	// Start Unix domain socket listener
	go startUnixDomainServer()
	fmt.Println("Unix domain server launched...")

	// create the pages directory if it doesn't exist
	util.Create_dir_if_not_exists(globals.OS_page_path)

	loadPasswordsHashesFromFile()
	mux := http.NewServeMux()

	mux.HandleFunc(lock_path, lockpageHandler)
	mux.HandleFunc(login_path1, loginHandler)
	mux.HandleFunc(login_path2, loginHandler)
	mux.HandleFunc(signup_path1, signupHandler)
	mux.HandleFunc(signup_path2, signupHandler)
	mux.HandleFunc(md2html_path, md2htmlHandler)
	mux.HandleFunc(view_path, viewHandler)                                             // TODO: fix the viewHandler
	mux.HandleFunc(update_path, pagelib.CheckPageExistsWrapper(pagelib.UpdateHandler)) // we do the page lock check in the UpdateHandler itself.
	mux.HandleFunc(edit_path, pagelib.CheckPageLockedWrapper(editors.MarkitupEditorHandler))
	mux.HandleFunc(debug_path, pagelib.CheckPageLockedWrapper(editors.DebugEditorHandler))
	mux.HandleFunc("/", rootHandler)

	pubfs := http.FileServer(http.Dir("public_assets"))
	mux.Handle("/public_assets/", http.StripPrefix("/public_assets/", pubfs))

	privfs := http.FileServer(http.Dir("private_assets"))
	mux.Handle("/private_assets/", http.StripPrefix("/private_assets/", privfs))

	log.Fatal(http.ListenAndServe(":8080", loginchecker{mux}))
}
