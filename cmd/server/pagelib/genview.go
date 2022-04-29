package pagelib

import (
	"net/http"
)

func generateViewPage(r *http.Request, v *ViewPageStruct) []byte {
	body := []byte(`
	<h1>` + v.PageTitle + `</h1>

<p>[<a href="/edit/` + v.PageID + `">edit</a>] [<a href="/debug/` + v.PageID + `">edit using the debug editor</a>]</p>

<div>`)
	contents := v.HTML
	end_div_tag := []byte("</div>")

	var result []byte
	result = append(result, body...)
	result = append(result, contents...)
	result = append(result, end_div_tag...)
	return result
}
