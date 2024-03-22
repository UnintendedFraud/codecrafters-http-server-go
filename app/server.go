package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type RequestData struct {
	path      string
	userAgent string
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(con)
	}
}

func handleConnection(con net.Conn) {
	defer func() {
		if err := con.Close(); err != nil {
			fmt.Println("Error closing the connection: ", err.Error())
			os.Exit(1)
		}
	}()

	data, err := getRequestData(con)
	if err != nil {
		fmt.Println("Error reading the request data: ", err.Error())
		os.Exit(1)
	}

	con.Write([]byte(data.getResponse()))
}

func getRequestData(con net.Conn) (RequestData, error) {
	b := make([]byte, 256)

	_, err := con.Read(b)
	if err != nil {
		return RequestData{}, err
	}

	reqData := string(b)
	data := strings.Split(reqData, "\r\n")

	path := strings.Split(data[0], " ")[1]

	var userAgent string
	if path == "/user-agent" {
		userAgent = strings.TrimSpace(strings.Split(data[2], ":")[1])
	}

	return RequestData{
		path,
		userAgent,
	}, nil
}

func (rd RequestData) getResponse() string {
	if strings.Contains(rd.path, "echo") {
		str := strings.Split(rd.path, "echo/")[1]

		return fmt.Sprintf(
			`HTTP/1.1 200 OK
            Content-Type: text/plain
            Content-Length: %d

            %s`, len(str), str,
		)
	}

	if strings.Contains(rd.path, "user-agent") {
		return fmt.Sprintf(
			`HTTP/1.1 200 OK
            Content-Type: text/plain
            Content-Length: %d

            %s`, len(rd.userAgent), rd.userAgent,
		)
	}

	if rd.path == "/" {
		return "HTTP/1.1 200 OK\r\n\r\n"
	}

	return "HTTP/1.1 404 Not Found\r\n\r\n"
}
