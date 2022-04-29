// This is the markitup editor.
// It's currently the main editor.
package editors

import (
	"net/http"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
	"github.com/1f604/supersimplewiki/cmd/server/pagelib"
	"github.com/1f604/supersimplewiki/cmd/server/util"
)

func MarkitupEditorHandler(w http.ResponseWriter, r *http.Request, title string) {

	pagelib.UnifiedViewEditHandler(w, r, globals.ENUM_EDITPAGE)

	p, err := pagelib.LoadPage(title)
	if err != nil {
		util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
		return
	}
	pagelib.RenderHTMLPage(w, r, globals.ENUM_EDITPAGE, p)

	// TODO: Fix this.
	/*
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		w.Write(body)

		p, err := loadPage(title)
		if err != nil {
			p = &Page{Title: title}
		}
		renderTemplate(w, r, "edit", p)
	*/

}
