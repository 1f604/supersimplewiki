// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	static_word  = "public"
	login_word   = "login"
	signup_word  = "signup"
	os_page_path = "./pages/"
	view_path    = "/view/"
	edit_path    = "/edit/"
	update_path  = "/update/"
	lock_path    = "/lock_page/"
	login_path1  = "/" + login_word
	login_path2  = "/" + login_word + "/"
	signup_path1 = "/" + signup_word
	signup_path2 = "/" + signup_word + "/"
	static_path  = "/" + static_word + "/"
)

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, edit_path+title, http.StatusFound)
		return
	}
	renderTemplate(w, r, "view", p)
}

var validPath = regexp.MustCompile("^/(edit|update|view)/([0-9]+)$")

func wrapViewEditHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			writeHTTPNoRefreshResponse(w, 406, "Error: invalid URL format.")
			return
		}
		fn(w, r, m[2])
	}
}

func lockpageHandler(w http.ResponseWriter, r *http.Request) {
	pageid := r.URL.Path[len(lock_path):]
	extendEditLock(pageid, w, r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeHTTPNoRefreshResponse(w, 404, "404 page not found")
		return
	}
	writeHTTPNoRefreshResponse(w, 200, "This is the home page.")
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

	// make sure the pages directory exists
	newpath := filepath.Join(".", "pages")
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	loadPasswordsHashesFromFile()
	mux := http.NewServeMux()

	mux.HandleFunc(view_path, wrapViewEditHandler(viewHandler))
	mux.HandleFunc(edit_path, wrapViewEditHandler(editHandler))
	mux.HandleFunc(update_path, wrapViewEditHandler(updateHandler))
	mux.HandleFunc(lock_path, lockpageHandler)
	mux.HandleFunc(login_path1, loginHandler)
	mux.HandleFunc(login_path2, loginHandler)
	mux.HandleFunc(signup_path1, signupHandler)
	mux.HandleFunc(signup_path2, signupHandler)
	mux.HandleFunc("/", rootHandler)

	fs := http.FileServer(http.Dir(static_word))
	mux.Handle(static_path, http.StripPrefix(static_path, fs))

	log.Fatal(http.ListenAndServe(":8080", loginchecker{mux}))
}
