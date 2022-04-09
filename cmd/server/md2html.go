// This file contains functions relating to the
// API endpoint that takes POST requests containing Markdown and responds with rendered HTML
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

func markdownToHTMLhoedown(s []byte) []byte {
	cmd := exec.Command("../../dependencies/hoedownv3.08/hoedown", "--html", "--tables", "--fenced-code", "--hard-wrap", "--toc-level", "500")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		//io.WriteString(stdin, "values written to stdin are passed to cmd's standard input")
		stdin.Write(s)
	}()

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		if stdoutStderr == nil {
			println("hoedown: No output.")
		} else {
			fmt.Printf("hoedown: %s\n", stdoutStderr)
		}
		log.Fatal("hoedown error: ", err)
	}
	return stdoutStderr
}

func md2htmlHandler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	w.Write(markdownToHTMLhoedown(body))
}
