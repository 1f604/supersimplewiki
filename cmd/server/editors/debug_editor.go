// This is the debug editor.
// It's currently the backup editor, in case you want to use it.
package editors

import (
	"io/ioutil"
	"net/http"
)

func DebugEditorHandler(w http.ResponseWriter, req *http.Request, title string) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	w.Write(body)
}
