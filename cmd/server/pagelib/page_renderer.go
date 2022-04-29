// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pagelib

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
	util "github.com/1f604/supersimplewiki/cmd/server/util"
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

func LoadPageSource(pageID string) (*EditPageStruct, error) {
	filename := globals.OS_page_path + pageID + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := EditPageStruct{
		PageID:    pageID,
		Source:    body,
		PageTitle: pageID,
		Checksum:  util.GetSHA1sum(body),
	}
	return &result, nil
}

func LoadRenderedPage(pageID string) (*ViewPageStruct, error) {
	filename := globals.OS_page_path + pageID + ".html"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	result := ViewPageStruct{
		PageID:    pageID,
		HTML:      body,
		PageTitle: pageID,
	}
	return &result, nil
}

var pageTitleDict = map[int]string{
	globals.ENUM_EDITPAGE:  "Editing page ",
	globals.ENUM_DEBUGPAGE: "Editing page ",
	globals.ENUM_VIEWPAGE:  "Viewing page ",
}

func RenderHTMLPage(w http.ResponseWriter, r *http.Request, page_type int) {
	/*
		p2r := PageToRender{
			Title:    p.Title,
			Body:     p.Body,
			HTML:     p.HTML,
			Username: util.GetUsernameFromRequest(r),
			Checksum: ,
		}*/
	w.Header().Set("Content-Type", "text/html") // apparently this is required if your HTML is not valid.
	var result []byte
	var contents []byte

	headertags := generatePageHeadTags(pageTitleDict[page_type] + p.Title)
	header := generateLoggedInPageHeader(r)
	result = append(result, headertags...)
	result = append(result, header...)

	switch page_type {
	case globals.ENUM_VIEWPAGE:
		contents = RenderViewPage(w, r, p)
	case globals.ENUM_EDITPAGE:
		contents = RenderEditPage(w, r, p)
	case globals.ENUM_DEBUGPAGE:
		contents = RenderDebugPage(w, r, p)
	}

	result = append(result, contents...)

	w.Write(result)
}

func RenderViewPage(w http.ResponseWriter, r *http.Request, p *Page) []byte {
	vps := ViewPageStruct{
		PageID:    p.PageID,
		PageTitle: p.Title,
		HTML:      p.Body,
	}
	return generateViewPage(r, &vps)
}

func RenderDebugPage(w http.ResponseWriter, r *http.Request, p *Page) []byte {
	eps := EditPageStruct{
		PageID:    p.PageID,
		PageTitle: p.Title,
		Source:    p.Body,
		Checksum:  util.GetSHA1sum(p.Body),
	}
	return generateDebugEditPage(r, &eps)
}

func RenderEditPage(w http.ResponseWriter, r *http.Request, p *Page) []byte {
	return []byte("NULL")
}
