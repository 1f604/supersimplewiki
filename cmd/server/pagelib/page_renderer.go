// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pagelib

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
	"github.com/1f604/supersimplewiki/cmd/server/util"
)

func CheckPageExists(pageID string) bool {
	filename := globals.OS_page_path + pageID + ".md"

	if _, err := os.Stat(filename); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
		// path/to/whatever does *not* exist
	} else {
		log.Fatal("CheckPageExists unexpected error: ", err)
	}
	return false
}

func LoadPage(title string) (*Page, error) {
	filename := globals.OS_page_path + title + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, HTML: template.HTML(body)}, nil
}

var templates = template.Must(template.ParseFiles("private_assets/editors/debug/edit.html", "./private_assets/view.html"))

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
	Checksum string
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, p *Page) {
	p2r := PageToRender{
		Title:    p.Title,
		Body:     p.Body,
		HTML:     p.HTML,
		Username: util.GetUsernameFromRequest(r),
		Checksum: util.GetSHA1sum(p.Body),
	}
	w.Header().Set("Content-Type", "text/html") // apparently this is required if your HTML is not valid.
	err := templates.ExecuteTemplate(w, tmpl+".html", p2r)
	if err != nil {
		util.WriteHTTPNoRefreshResponse(w, http.StatusInternalServerError, err.Error())
	}
}
