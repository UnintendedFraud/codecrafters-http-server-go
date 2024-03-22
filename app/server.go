package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type RequestData struct {
	path string
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

	con.Write([]byte(getResponse(data.path)))
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

	return RequestData{
		path,
	}, nil
}

func getResponse(path string) string {
	if strings.Contains(path, "echo") {
		str := strings.Split(path, "echo/")[1]

		return fmt.Sprintf(
			`HTTP/1.1 200 OK
Context-Type: text/plain
Content-Length: %d

%s`,
			len(str),
			str,
		)
	}
	if path == "/" {
		return "HTTP/1.1 200 OK\r\n\r\n"
	}

	return "HTTP/1.1 404 Not Found\r\n\r\n"
}
