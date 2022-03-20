// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type ErrLocked struct{}

func (m ErrLocked) Error() string {
	return "Error: Someone else is already editing the page"
}

var page_edit_lock = make(map[string]*uint32)

func tryGrabLockOnPage(pageid string) error {
	ptr, ok := page_edit_lock[pageid]
	if !ok {
		ptr = new(uint32)
		page_edit_lock[pageid] = ptr
	}

	if !atomic.CompareAndSwapUint32(ptr, 0, 1) {
		return ErrLocked{}
	}
	//	defer atomic.StoreUint32(ptr, 0)

	return nil
}

func releaseLockOnPage(pageid string) {
	ptr, ok := page_edit_lock[pageid]
	if !ok {
		log.Fatal("page ID not in page lock map")
	}

	atomic.StoreUint32(ptr, 0)
}

type Page struct {
	Title string
	Body  []byte
	HTML  template.HTML
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, HTML: template.HTML(body)}, nil
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
