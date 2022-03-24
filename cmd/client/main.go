package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

const (
	unixDomainProtocol = "unix"
	unixSockAddr       = "/tmp/supersimplewiki.sock"
)

// credits to https://github.com/devlights/go-unix-domain-socket-example
func sendMsgToServer(msg string) {
	conn, err := net.Dial(unixDomainProtocol, unixSockAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Fatal(err)
	}

	err = conn.(*net.UnixConn).CloseWrite()
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("supersimplewiki client")
	fmt.Println("-----------------------")

	for {
		fmt.Print("> ")
		msg, _ := reader.ReadString('\n')
		msg = strings.Replace(msg, "\n", "", -1)

		sendMsgToServer(msg)
	}
}
