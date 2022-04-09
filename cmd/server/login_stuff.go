package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var passwordRegex = regexp.MustCompile(`^(?P<Username>[a-z0-9_]+) (?P<PwdHash>[a-zA-Z0-9+/=]+) activated:(?P<Activated>true|false) birthyear:(?P<BirthYear>[0-9]+) realname:(?P<RealName>[a-zA-Z ]+)$`)

type UserInfo struct {
	PwdHash   string
	Activated bool
	RealName  string
	BirthYear string
}

// if you want to "reset" a password, just delete the line from the passwords.txt file and restart the server
// then sign up again with that username. It will just change the password to the new one.
var password_file_path = "passwords.txt" // this file is the single source of truth for user accounts
var noSubmitOnRefreshJS = `<script src="/` + static_word + `/norefresh.js"></script>
`

// should probably use argon2 for this...but I don't want to import any external libraries
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	str := base64.StdEncoding.EncodeToString(sum[:])
	return str
}

var userInfoMap = map[string]*UserInfo{} //maps usernames to account info
var tokenMap = map[string]string{}       //maps session tokens to usernames

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

// Currently, you can activate a user in 2 ways:
// 1. Update the passwords.txt file manually. You don't have to restart the server.
// 2. Use the command line client. You don't have to restart the server.
func activateUser(ptr *UserInfo, username_to_activate string) bool {
	ptr.Activated = true

	// now find the user's record in the file and update it to say that the user is activated
	// just read the whole contents of the file, find the line, and write it back
	input, err := ioutil.ReadFile(password_file_path)
	if err != nil { // this should never happen
		log.Println("ERROR: Failed to open password file to activate user.")
		return false
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		// skip empty lines
		if len(line) < 2 {
			continue
		}
		if !passwordRegex.Match([]byte(line)) { // this should never happen
			log.Println("ERROR: Line in password file failed to match regex to activate user.")
			return false
		}
		capturedGroups := passwordRegex.FindStringSubmatch(line)
		if len(capturedGroups) != 6 {
			log.Fatal("Unexpected password file format.")
		}
		username := capturedGroups[1]
		if username != username_to_activate {
			continue // go to next line
		} else { // update the line
			pwdhash := capturedGroups[2]
			birthyear := capturedGroups[4]
			realname := capturedGroups[5]
			newline := username + " " + pwdhash + " activated:true birthyear:" + birthyear + " realname:" + realname // + "\n"
			lines[i] = newline
			goto userFound
		}
	}
	log.Println("ERROR: Failed to find user in password file.")
	return false

userFound: // now write the new contents back to the file
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(password_file_path, []byte(output), 0600)
	if err != nil {
		log.Println("ERROR: Failed to update password file with activated user.")
		return false
	}

	return true
}

func loadPasswordsHashesFromFile() {
	f, err := os.OpenFile(password_file_path, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	userInfoMap = map[string]*UserInfo{} // clear the in-memory map every time someone signs up
	fscanner := bufio.NewScanner(f)
	for fscanner.Scan() {
		line := fscanner.Text()
		// ignore empty lines
		if len(line) < 2 {
			continue
		}
		if !passwordRegex.Match([]byte(line)) {
			log.Fatal("Failed to match password regex")
		}
		capturedGroups := passwordRegex.FindStringSubmatch(line)
		if len(capturedGroups) != 6 {
			log.Fatal("Unexpected password file format.")
		}
		username := capturedGroups[1]
		pwdhash := capturedGroups[2]
		activated := capturedGroups[3] == "true" // the regex only allows "true" or "false"
		birthyear := capturedGroups[4]
		realname := capturedGroups[5]

		userInfoMap[username] = &UserInfo{
			PwdHash:   pwdhash,
			RealName:  realname,
			BirthYear: birthyear,
			Activated: activated,
		}
	}
}

func doCreateNewAccount(username string, password string, realname string, birthyear string) {
	passwordhash := hashPassword(password)
	userInfoMap[username] = &UserInfo{
		PwdHash:   passwordhash,
		Activated: false,
		RealName:  realname,
		BirthYear: birthyear,
	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(password_file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	line := username + " " + passwordhash + " activated:false birthyear:" + birthyear + " realname:" + realname + "\n"
	f.Write([]byte(line))
}

func isUsernameValid(u string) bool {
	usernameRegex := regexp.MustCompile(`^[a-z0-9_]+$`)
	return usernameRegex.MatchString(u)
}

func isRealnameValid(u string) bool {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z ]+$`)
	return usernameRegex.MatchString(u)
}

func isBirthyearValid(u string) bool {
	usernameRegex := regexp.MustCompile(`^\d{4}$`)
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

		username := r.FormValue("username")
		password1 := r.FormValue("password1")
		password2 := r.FormValue("password2")
		realname := r.FormValue("realname")
		birthyear := r.FormValue("birthyear")

		// check if username already exists
		loadPasswordsHashesFromFile() // clear the in memory map and reload from file
		_, ok := userInfoMap[username]
		if ok {
			writeHTTPNoRefreshResponse(w, 400, "Error: username is already registered.")
			return
		}
		// check passwords match
		if password1 != password2 {
			writeHTTPNoRefreshResponse(w, 400, "Error: passwords do not match.")
			return
		}
		// check username requirements
		if !isUsernameValid(username) {
			writeHTTPNoRefreshResponse(w, 400, "Error: username contains invalid characters. Only numbers, letters, and underscore is allowed.")
			return
		}
		// check name requirements
		if !isRealnameValid(realname) {
			writeHTTPNoRefreshResponse(w, 400, "Error: real name contains invalid characters. Only letters and spaces are allowed.")
			return
		}
		// check year requirements
		if !isBirthyearValid(birthyear) {
			writeHTTPNoRefreshResponse(w, 400, "Error: invalid birth year.")
			return
		}
		// check length requirements
		if len(username) < 2 {
			writeHTTPNoRefreshResponse(w, 400, "Please enter a username of at least length 2.")
			return
		}
		if len(password1) < 2 {
			writeHTTPNoRefreshResponse(w, 400, "Please enter a password of at least length 2.")
			return
		}
		if len(realname) < 2 {
			writeHTTPNoRefreshResponse(w, 400, "Please enter a real name of at least length 2.")
			return
		}
		doCreateNewAccount(username, password1, realname, birthyear)
		writeHTTPNoRefreshResponse(w, 200, "Success! Now you can <a href=\"/login/\">log in</a> with your new account.")

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

}

func writeLoggedIn(w http.ResponseWriter, r *http.Request) {
	username := getUsernameFromRequest(r)
	msgBody := "You are already logged in. You are currently logged in as: " + username + "</br>Return to the <a href=\"/\">home page</a>."
	writeHTTPNoRefreshResponse(w, 200, msgBody)
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

		loadPasswordsHashesFromFile() // clear the in memory map and reload from file
		// Get the expected password from our in memory map
		// Check user exists
		storedUserInfo, ok := userInfoMap[username]
		if !ok {
			writeHTTPNoRefreshResponse(w, http.StatusUnauthorized, "Login failed. Username does not exist.")
			return
		}

		// Check user's password matches what we have stored
		expectedPasswordHash := storedUserInfo.PwdHash
		actualPasswordHash := hashPassword(password)
		if expectedPasswordHash != actualPasswordHash {
			writeHTTPNoRefreshResponse(w, http.StatusUnauthorized, "Login failed. Wrong password.")
			return
		}

		// Check if user account is activated
		if !storedUserInfo.Activated {
			writeHTTPNoRefreshResponse(w, http.StatusUnauthorized, "Login failed. Account not activated.")
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
	_, ok := noAuthNeededWhitelist[firstpart]
	if !ok {
		if !validateAuth(r) { // check cookie
			writeHTTPNoRefreshResponse(w, http.StatusUnauthorized, "Not authorized. Please <a href=\"/login\">log in</a> or <a href=\"/signup\">sign up</a>.")
			return
		}
	}
	c.h.ServeHTTP(w, r)
}
