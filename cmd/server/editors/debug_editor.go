// This is the debug editor.
// It's currently the backup editor, in case you want to use it.
package editors

import (
	"net/http"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
	"github.com/1f604/supersimplewiki/cmd/server/pagelib"
	"github.com/1f604/supersimplewiki/cmd/server/util"
)

func DebugEditorHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := pagelib.LoadPage(title)
	if err != nil {
		util.WriteHTTPNoRefreshResponse(w, 404, "Error: wiki page ID not found.")
		return
	}
	pagelib.RenderHTMLPage(w, r, globals.ENUM_DEBUGPAGE, p)
}
