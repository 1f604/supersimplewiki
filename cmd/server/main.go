// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

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

func lockpageHandler(w http.ResponseWriter, r *http.Request) {
	pageid := r.URL.Path[len(lock_path):]
	pagelib.ExtendEditLock(pageid, w, r)
}

const (
	lNet  = "tcp"
	lAddr = ":12345"
)

func main() {
	_, err := net.Listen(lNet, lAddr)
	if err != nil {
		fmt.Println("Another instance of supersimplewiki is currently running!")
		return
	}

	// load edit page scripts
	pagelib.EditPageScriptsHTML, err = ioutil.ReadFile("./internal_assets/editPageScripts.html")
	if err != nil {
		log.Fatal("Failed to read edit page scripts: ", err)
	}

	// Start Unix domain socket listener
	go startUnixDomainServer()
	fmt.Println("Unix domain server launched...")

	// create the pages directory if it doesn't exist
	util.Create_dir_if_not_exists(globals.OS_page_path)

	loadPasswordsHashesFromFile()
	mux := http.NewServeMux()

	// handle special pages/API endpoints e.g. login, lock, and update API
	mux.HandleFunc(lock_path, lockpageHandler)
	mux.HandleFunc(login_path1, loginHandler)
	mux.HandleFunc(login_path2, loginHandler)
	mux.HandleFunc(signup_path1, signupHandler)
	mux.HandleFunc(signup_path2, signupHandler)
	mux.HandleFunc(md2html_path, md2htmlHandler)
	mux.HandleFunc(update_path, pagelib.CheckPageExistsWrapper(pagelib.UpdateHandler)) // we do the page lock check in the UpdateHandler itself.

	// serve public pages and assets
	pubfs := http.FileServer(http.Dir("public_assets"))
	mux.Handle("/public_assets/", http.StripPrefix("/public_assets/", pubfs))

	// handle normal pages e.g. view and edit
	mux.HandleFunc("/", UnifiedPageHandler)

	log.Fatal(http.ListenAndServe(":8080", loginchecker{mux}))
}
