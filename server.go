package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs !!!")

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
	// read the request data
	req := make([]byte, 1024)
	_, err := conn.Read(req)
	if err != nil {
		fmt.Println("Error Reading request: ", err.Error())
		conn.Close()
		return
	}

	//
	reqSplit := strings.Split(string(req), "\r\n")
	// for i := 0; i < len(reqSplit); i++ {
	// 	fmt.Printf("\n %d : %s", i, reqSplit[i])
	// }
	data := strings.Split(reqSplit[0], " ")
	reqMethod := data[0]
	path := data[1]

	resp := "HTTP/1.1 404 Unkown Method\r\n\r\n"
	fmt.Printf("\nRecieved a %s request on Path %s", reqMethod, path)
	if reqMethod == "GET" {
		resp = get(reqSplit, path)
	} else if reqMethod == "POST" {
		resp = post(reqSplit, path)
	}
	conn.Write([]byte(resp))
	conn.Close()

}

func get(reqSplit []string, path string) string {
	if path == "/" {
		return "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/echo/") {
		breakdown := strings.Split(string(path), "/")
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(breakdown[2]), breakdown[2])
	} else if strings.HasPrefix(string(path), "/user-agent") {
		userAgent := strings.Split(string(reqSplit[2]), " ")
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent[1]), userAgent[1])
	} else if strings.HasPrefix(string(path), "/files/") {
		dir := os.Args[2]
		breakdown := strings.Split(string(path), "/")
		content, err := os.ReadFile(dir + breakdown[2])
		if err != nil {
			fmt.Println(err.Error())
			return "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		// fmt.Println(content)
		return fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)
	}
	return "HTTP/1.1 404 Not Found\r\n\r\n"
}

func post(reqSplit []string, path string) string {
	if strings.HasPrefix(string(path), "/files/") {
		dir := os.Args[2]
		breakdown := strings.Split(string(path), "/")
		file := []byte(strings.Trim(reqSplit[len(reqSplit)-1], "\x00"))
		err := os.WriteFile(dir+breakdown[2], file, os.ModePerm)
		if err != nil {
			fmt.Println(err.Error())
			return "HTTP/1.1 404 Write failed\r\n\r\n"
		}
		// fmt.Println(content)
		return "HTTP/1.1 201 Created\r\n\r\n"
	}
	return "HTTP/1.1 404 Not Found\r\n\r\n"
}
