package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
)

const (
	unixDomainProtocol = "unix"
	unixSockAddr       = "/tmp/supersimplewiki.sock"
)

func cleanupUnixDomainSocket() {
	if _, err := os.Stat(unixSockAddr); err == nil {
		if err := os.RemoveAll(unixSockAddr); err != nil {
			log.Fatal(err)
		}
	}
}

// Currently all this server does is to listen to activate user requests
// But you can already do that by editing the passwords file directly
// In future I will add more commands to make this server more useful.
func startUnixDomainServer() {
	cleanupUnixDomainSocket()
	UnixListener, err := net.Listen(unixDomainProtocol, unixSockAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer UnixListener.Close()

	for {
		conn, err := UnixListener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleClientMsg(conn)
	}
}

var activationRequestRegex = regexp.MustCompile(`^activate ([a-z0-9_]+)$`)

func handleClientMsg(conn net.Conn) {
	defer conn.Close()
	log.Printf("Received new message from client\n")

	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, conn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(">>> ", buf.String())
	var response string
	var capturedGroups []string
	var ptr *UserInfo
	var ok bool

	if !activationRequestRegex.Match(buf.Bytes()) {
		response = "Failed to match activation request regex. The syntax is: activate username"
		goto sendResponse
	}
	capturedGroups = activationRequestRegex.FindStringSubmatch(buf.String())

	if len(capturedGroups) != 2 {
		response = "Invalid activation request"
		goto sendResponse
	}

	// check if username actually exists
	ptr, ok = userInfoMap[capturedGroups[1]]
	if !ok {
		response = "Error: username " + capturedGroups[1] + " not found."
	} else {
		ok = activateUser(ptr, capturedGroups[1])
		if ok {
			response = "Activation request successful. User " + capturedGroups[1] + " is now activated."
		} else {
			response = "Failed to update user record in file. This should never happen."
		}
	}

sendResponse:
	buf.Reset()
	buf.WriteString(response)
	_, err = io.Copy(conn, buf)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Sent response:", response)
}
