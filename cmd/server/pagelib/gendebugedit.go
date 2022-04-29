package pagelib

import (
	"html"
	"net/http"
)

var EditPageScriptsHTML []byte

func generateDebugEditPage(r *http.Request, v *EditPageStruct) []byte {
	scripts := EditPageScriptsHTML
	escapedSource := html.EscapeString(string(v.Source))
	var result []byte

	result = append(result, []byte(`<script>
	var g_pageID = "`+v.PageID+`"
	</script>
	`)...)

	result = append(result, scripts...)

	result = append(result, []byte(`
	<h1>Editing `+v.PageTitle+`</h1>
	<div><textarea id="bodyfield" rows="20" cols="80">`)...)

	result = append(result, []byte(escapedSource)...)

	result = append(result, []byte(`</textarea></div>
	<div><input type="hidden" id="hidden_csum" value="`+v.Checksum+`" /></div>
	<div id = "displayErrorMsgDiv"></div>
	<div> <button id = "btnSubmit" type="button">Update</button></div>`)...)

	return result
}
