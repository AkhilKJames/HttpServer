package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type HTTPRequest struct {
	method      string
	target      string
	httpVersion string
	headers     map[string]string
	body        []byte
}

type HTTPResponse struct {
	statusCode  int16
	status      string
	httpVersion string
	headers     map[string]string
	body        string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Creates a port listener
	l, err := net.Listen("tcp", "127.0.0.1:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		// accepts incoming request
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			conn.Close()
			return
		}
		go handleConnections(conn)
	}
}

func handleConnections(conn net.Conn) {
	// fmt.Println("------------------------------ New Connection ")
	// fmt.Println("Connected: ", conn.RemoteAddr())
	request, err := parseRequest(conn)
	if err != nil {
		fmt.Println("Failed to parse request: " + err.Error())
	}

	// fmt.Println("--------------------")
	fmt.Println(request)

	resp := HTTPResponse{}
	if request.method == "GET" {
		resp = get(request, request.target)
	} else if request.method == "POST" {
		resp = post(request, request.target)
	} else {
		resp.statusCode = 404
		resp.httpVersion = request.httpVersion
		resp.status = "Not Found"
	}
	conn.Write([]byte(responseWriter(resp)))
	conn.Close()
}

func parseRequest(connection net.Conn) (HTTPRequest, error) {
	lines := make(chan []byte)
	go byteReader(lines, connection)
	requestLineValues := bytes.Split(<-lines, []byte(" "))
	if len(requestLineValues) != 3 {
		return HTTPRequest{}, fmt.Errorf("invalid request line")
	}
	method := string(requestLineValues[0])
	target := string(requestLineValues[1])
	httpVersion := string(requestLineValues[2])
	headers := make(map[string]string)
	for {
		headerLine := <-lines
		if len(headerLine) == 0 {
			break
		}
		headerLineValues := bytes.Split(headerLine, []byte(": "))
		headers[string(headerLineValues[0])] = string(headerLineValues[1])
	}
	body := <-lines
	return HTTPRequest{
		method:      method,
		target:      target,
		httpVersion: httpVersion,
		headers:     headers,
		body:        body,
	}, nil
}

func byteReader(channel chan []byte, connection net.Conn) error {
	buffer := make([]byte, 0)
	for {
		tmp := make([]byte, 1024)
		n, err := connection.Read(tmp)
		if err != nil && err != io.EOF {
			close(channel)
			return err
		}
		buffer = append(buffer, tmp[:n]...)
		splitBuffer := bytes.Split(buffer, []byte("\r\n"))
		for _, line := range splitBuffer[:len(splitBuffer)-1] {
			channel <- line
		}
		buffer = splitBuffer[len(splitBuffer)-1]
		if n <= len(tmp) || err == io.EOF {
			break
		}
	}
	channel <- buffer
	close(channel)
	return nil
}

func get(req HTTPRequest, path string) HTTPResponse {
	if path == "/" {
		return HTTPResponse{
			statusCode:  200,
			status:      "OK",
			httpVersion: req.httpVersion,
		}
	} else if strings.HasPrefix(path, "/echo/") {
		breakdown := strings.Split(string(path), "/")
		return HTTPResponse{
			statusCode:  200,
			status:      "OK",
			httpVersion: req.httpVersion,
			headers:     map[string]string{"Content-Type": req.headers["Content-Type"], "Content-Length": strconv.Itoa(len(breakdown[2]))},
			body:        breakdown[2],
		}
	} else if strings.HasPrefix(string(path), "/user-agent") {
		userAgent := req.headers["User-Agent"]
		return HTTPResponse{
			statusCode:  200,
			status:      "OK",
			httpVersion: req.httpVersion,
			headers:     map[string]string{"Content-Type": req.headers["Content-Type"], "Content-Length": strconv.Itoa(len(userAgent))},
			body:        userAgent,
		}
	} else if strings.HasPrefix(string(path), "/files/") {
		dir := os.Args[2]
		breakdown := strings.Split(string(path), "/")
		content, err := os.ReadFile(dir + breakdown[2])
		if err != nil {
			fmt.Println(err.Error())
			return HTTPResponse{
				statusCode:  404,
				status:      "Not Found",
				httpVersion: req.httpVersion,
			}
		}
		return HTTPResponse{
			statusCode:  200,
			status:      "OK",
			httpVersion: req.httpVersion,
			headers:     map[string]string{"Content-Type": req.headers["Content-Type"], "Content-Length": strconv.Itoa(len(content))},
			body:        string(content),
		}
	}
	return HTTPResponse{
		statusCode:  404,
		status:      "Not Found",
		httpVersion: req.httpVersion,
	}
}

func post(req HTTPRequest, path string) HTTPResponse {
	if strings.HasPrefix(string(path), "/files/") {
		dir := os.Args[2]
		breakdown := strings.Split(string(path), "/")
		err := os.WriteFile(dir+breakdown[2], req.body, os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return HTTPResponse{
				statusCode:  404,
				status:      "Write Failed",
				httpVersion: req.httpVersion,
			}
		}
		return HTTPResponse{
			statusCode:  201,
			status:      "Created",
			httpVersion: req.httpVersion,
		}
	}
	return HTTPResponse{
		statusCode:  404,
		status:      "Not Found",
		httpVersion: req.httpVersion,
	}
}

func responseWriter(resp HTTPResponse) string {
	res := fmt.Sprintf("%s %d %s\r\n", resp.httpVersion, resp.statusCode, resp.status)
	if resp.body != "" {
		var header string
		for key, val := range resp.headers {
			header += fmt.Sprintf("%s: %s\r\n", key, val)
		}
		res = res + header + fmt.Sprintf("\r\n%s", string(resp.body))
	}
	return res
}
