package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net"
	"os"
	"strings"
)

type RequestData struct {
	path      string
	method    string
	userAgent string
}

func main() {
	fDirectory := flag.String("directory", "", "directory for file endpoint")
	flag.Parse()

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

		go handleConnection(con, *fDirectory)
	}
}

func handleConnection(con net.Conn, dir string) {
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

	con.Write([]byte(data.getResponse(dir)))
}

func getRequestData(con net.Conn) (RequestData, error) {
	b := make([]byte, 256)

	_, err := con.Read(b)
	if err != nil {
		return RequestData{}, err
	}

	reqData := string(b)
	data := strings.Split(reqData, "\r\n")

	split := strings.Split(data[0], " ")
	method := split[0]
	path := split[1]

	var userAgent string
	if path == "/user-agent" {
		userAgent = strings.TrimSpace(strings.Split(data[2], ":")[1])
	}

	return RequestData{
		path,
		method,
		userAgent,
	}, nil
}

func (rd RequestData) getResponse(dir string) string {
	if strings.HasPrefix(rd.path, "/echo") {
		str := strings.Split(rd.path, "echo/")[1]

		return Content("text/plain", len(str), str)
	}

	if strings.HasPrefix(rd.path, "/user-agent") {
		return Content("text/plain", len(rd.userAgent), rd.userAgent)
	}

	if strings.HasPrefix(rd.path, "/files") {
		filename := strings.Split(rd.path, "files/")[1]
		fPath := dir + filename

		fContent, err := os.ReadFile(fPath)
		if err != nil {
			fmt.Println("Error reading file: ", fPath, err.Error())
			return NotFound()
		}

		if rd.method == "POST" {
			if err = os.WriteFile(fPath, fContent, fs.ModeDevice); err != nil {
				fmt.Println("Error writing file: ", fPath, err.Error())
				panic(err)
			}

			return Created()
		}

		return Content("application/octet-stream", len(fContent), string(fContent))
	}

	if rd.path == "/" {
		return "HTTP/1.1 200 OK\r\n\r\n"
	}

	return NotFound()
}

func NotFound() string {
	return "HTTP/1.1 404 Not Found\r\n\r\n"
}

func Created() string {
	return "HTTP/1.1 201 Created\r\n\r\n"
}

func Content(
	contentType string,
	contentLength int,
	content string,
) string {
	return fmt.Sprintf(
		`HTTP/1.1 200 OK
Content-Type: %s
Content-Length: %d

%s`,
		contentType,
		contentLength,
		content,
	)
}
