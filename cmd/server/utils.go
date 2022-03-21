package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"os"
)

func writeHeaderNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

func writeBodyNoRefresh(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(noSubmitOnRefreshJS))
	w.Write(bytes)
}

// used for creating new session tokens
func getRandomStringBASE64() string {
	c := 24
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("error:", err)
	}
	// The slice should now contain random bytes instead of only zeroes.
	str := base64.StdEncoding.EncodeToString(b)
	return str
}

// used for creating new filenames
func getRandomStringHex(c int) string {
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("error:", err)
	}
	// The slice should now contain random bytes instead of only zeroes.
	str := hex.EncodeToString(b)
	return str
}

// Guaranteed to return a filename that does not currently exist
func generateRandomFilename() string {
	for {
		name := getRandomStringHex(5) // 255 ^ 5 possible file names.
		if _, err := os.Stat("./pages/" + name); errors.Is(err, os.ErrNotExist) {
			return name
		}
		//println("Oh no! Filename already exists!")
	}
}
