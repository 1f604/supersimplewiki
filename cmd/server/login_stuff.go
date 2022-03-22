package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// if you want to "reset" a password, just delete the line from the passwords.txt file and restart the server
// then sign up again with that username. It will just change the password to the new one.
var password_file_path = "passwords.txt" // this file is the single source of truth for user accounts
var noSubmitOnRefreshJS = `<script src="/` + static_word + `/norefresh.js"></script>`

// should probably use argon2 for this...but I don't want to import any external libraries
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	str := base64.StdEncoding.EncodeToString(sum[:])
	return str
}

var passwordHashMap = map[string]string{} //maps usernames to passwords
var tokenMap = map[string]string{}        //maps session tokens to usernames

type loginchecker struct {
	h http.Handler
}

// Checks whether the user is logged in
func validateAuth(r *http.Request) bool {
	c, err := r.Cookie("session_token")
	if err != nil {
		return false
	}
	sessionToken := c.Value
	_, ok := tokenMap[sessionToken]
	return ok
}

// This function assumes the user is already logged in.
func getUsernameFromRequest(r *http.Request) string {
	c, err := r.Cookie("session_token")
	if err != nil {
		log.Fatal("Failed to get session token")
	}
	sessionToken := c.Value
	username, ok := tokenMap[sessionToken]
	if !ok {
		log.Fatal("Failed to get username from session token")
	}
	return username
}

func loadPasswordsHashesFromFile() {
	f, err := os.OpenFile(password_file_path, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fscanner := bufio.NewScanner(f)
	for fscanner.Scan() {
		line := fscanner.Text()
		words := strings.Fields(line)
		if len(words) != 2 {
			log.Fatal("Unexpected password file format.")
		}
		username := words[0]
		pwdhash := words[1]
		passwordHashMap[username] = pwdhash
	}
}

func storeNewPasswordInFile(username string, passwordhash string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(password_file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write([]byte(username + " " + passwordhash + "\n"))
}

func doCreateNewAccount(username string, password string) {
	passwordhash := hashPassword(password)
	passwordHashMap[username] = passwordhash
	storeNewPasswordInFile(username, passwordhash)
}

func isUsernameValid(u string) bool {
	usernameRegex := regexp.MustCompile(`^[a-z0-9_]+$`)
	return usernameRegex.MatchString(u)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if validateAuth(r) { // if already logged in
		writeLoggedIn(w, r)
		return
	}

	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "signup.html")
	case "POST":
		// parse the form
		if err := r.ParseForm(); err != nil {
			log.Fatal("ParseForm() err: " + err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(noSubmitOnRefreshJS))

		username := r.FormValue("username")
		password1 := r.FormValue("password1")
		password2 := r.FormValue("password2")

		// check if username already exists
		_, ok := passwordHashMap[username]
		if ok {
			w.Write([]byte("Error: username is already registered."))
			return
		}
		// check passwords match
		if password1 != password2 {
			w.Write([]byte("Error: passwords do not match."))
			return
		}
		// check username requirements
		if !isUsernameValid(username) {
			w.Write([]byte("Error: username contains invalid characters. Only numbers, letters, and underscore is allowed."))
			return
		}
		// check length requirements
		if len(username) < 2 {
			w.Write([]byte("Please enter a username of at least length 2."))
			return
		}
		if len(password1) < 2 {
			w.Write([]byte("Please enter a password of at least length 2."))
			return
		}
		doCreateNewAccount(username, password1)
		w.Write([]byte("Success! Now you can <a href=\"/login/\">log in</a> with your new account."))

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

}

func writeLoggedIn(w http.ResponseWriter, r *http.Request) {
	username := getUsernameFromRequest(r)
	writeBodyNoRefresh(w, []byte("You are already logged in. You are currently logged in as: "+username))
	w.Write([]byte("</br>Return to the <a href=\"/\">home page</a>."))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if validateAuth(r) { // if already logged in
		writeLoggedIn(w, r)
		return
	}

	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "login.html")
	case "POST":
		// parse the form
		if err := r.ParseForm(); err != nil {
			log.Fatal("ParseForm() err: " + err.Error())
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Get the expected password from our in memory map
		expectedPasswordHash, ok := passwordHashMap[username]
		actualPasswordHash := hashPassword(password)

		// Check if password hash matches what we have stored
		if !ok || expectedPasswordHash != actualPasswordHash {
			w.WriteHeader(http.StatusUnauthorized)
			writeBodyNoRefresh(w, []byte("Login failed. Wrong username/password."))
			return
		}

		// Create a new random session token
		sessionToken := getRandomStringBASE64()

		// Save the token in the session map
		tokenMap[sessionToken] = username

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().AddDate(1000, 0, 0), // expires 1000 years in the future
			Path:    "/",                            // we want this cookie to be sent along with every request
		})

		writeBodyNoRefresh(w, []byte("Login successful! Return to the <a href=\"/\">home page</a>."))

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

var noAuthNeededWhitelist = map[string]bool{
	static_word: true,
	login_word:  true,
	signup_word: true,
}

func (c loginchecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	firstpart := strings.Split(r.URL.Path, "/")[1]
	writeHeaderNoCache(w)
	if !noAuthNeededWhitelist[firstpart] {
		if !validateAuth(r) { // check cookie
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Not authorized. Please <a href=\"/login\">log in</a> or <a href=\"/signup\">sign up</a>."))
			return
		}
	}
	c.h.ServeHTTP(w, r)
}
