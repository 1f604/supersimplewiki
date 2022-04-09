package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/1f604/supersimplewiki/cmd/server/globals"
)

func WriteHeaderNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.
}

func WriteBodyNoRefresh(w http.ResponseWriter, bytes []byte) {
	const noSubmitOnRefreshJS = `<script src="/public_assets/norefresh.js"></script>
	`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(noSubmitOnRefreshJS))
	w.Write(bytes)
}

func WriteHTTPResponse(w http.ResponseWriter, errorcode int, msg string) {
	w.WriteHeader(errorcode)
	w.Write([]byte(msg))
}

func WriteHTTPNoRefreshResponse(w http.ResponseWriter, errorcode int, msg string) {
	w.WriteHeader(errorcode)
	WriteBodyNoRefresh(w, []byte(msg))
}

// used for creating new session tokens
func GetRandomStringBASE64() string {
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

// used for checking if text was modified
func GetSHA1sum(bytes []byte) string {
	h := sha1.New()
	h.Write(bytes)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

// used to check if file was modified while user was editing it
func GetSHA1sumOfFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal("Failed to get checksum of file:", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
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

// This function assumes the user is already logged in.
func GetUsernameFromRequest(r *http.Request) string {
	c, err := r.Cookie("session_token")
	if err != nil {
		log.Fatal("Failed to get session token")
	}
	sessionToken := c.Value
	username, ok := globals.TokenMap[sessionToken]
	if !ok {
		log.Fatal("Failed to get username from session token")
	}
	return username
}
