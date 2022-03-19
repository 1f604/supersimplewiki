// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"regexp"
)

const (
	view_path = "/view/"
	edit_path = "/edit/"
	save_path = "/save/"
)

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, edit_path+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func main() {
	http.HandleFunc(view_path, wrapViewEditHandler(viewHandler))
	http.HandleFunc(edit_path, wrapViewEditHandler(editHandler))
	http.HandleFunc(save_path, wrapViewEditHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
