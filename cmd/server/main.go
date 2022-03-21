// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
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
	save_path    = "/save/"
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

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	releaseEditLock(title)
	http.Redirect(w, r, view_path+title, http.StatusFound)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([0-9]+)$")

func wrapViewEditHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			w.WriteHeader(400)
			w.Write([]byte("Error: invalid URL format."))
			return
		}
		fn(w, r, m[2])
	}
}

// TODO: Add a "page is being edited by user x" message.
func lockpageHandler(w http.ResponseWriter, r *http.Request) {
	pageid := r.URL.Path[len(lock_path):]
	extendEditLock(pageid)
	w.WriteHeader(200)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))
		return
	}
	w.Write([]byte("This is the home page."))
}

func main() {
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
	mux.HandleFunc(save_path, wrapViewEditHandler(saveHandler))
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
