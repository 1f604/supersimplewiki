//taken from: https://go.dev/doc/articles/wiki/

package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	view_path = "/view/"
	edit_path = "/edit/"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[:])
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len(view_path):]
	fmt.Fprintf(w, "Hi there! I like %s!", title)
}

func main() {
	http.HandleFunc(view_path, viewHandler)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
