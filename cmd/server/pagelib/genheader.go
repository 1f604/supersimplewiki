package pagelib

import (
	"net/http"

	"github.com/1f604/supersimplewiki/cmd/server/util"
)

func generatePageHeadTags(pagetitle string) []byte {
	html := `<head>
			 <title>` + pagetitle + `</title> 
			 </head>`
	return []byte(html)
}

// returns the logged in page header HTML
// not very performant, fix if it becomes a problem
func generateLoggedInPageHeader(r *http.Request) []byte {
	username := util.GetUsernameFromRequest(r)
	html := `<p>You are currently logged in as: ` + username + `</p>`
	return []byte(html)
}
