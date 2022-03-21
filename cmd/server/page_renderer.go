// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"net/http"
	"os"
)

func (p *Page) save() error {
	filename := os_page_path + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := os_page_path + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, HTML: template.HTML(body)}, nil
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

type Page struct {
	Title string
	Body  []byte
	HTML  template.HTML
}

type PageToRender struct {
	Title    string
	Body     []byte
	HTML     template.HTML
	Username string
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, p *Page) {
	p2r := PageToRender{
		Title:    p.Title,
		Body:     p.Body,
		HTML:     p.HTML,
		Username: getUsernameFromRequest(r),
	}
	err := templates.ExecuteTemplate(w, tmpl+".html", p2r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
