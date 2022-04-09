// This is the markitup editor.
// It's currently the main editor.
package editors

import (
	"net/http"
)

func MarkitupEditorHandler(w http.ResponseWriter, req *http.Request, title string) {

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
