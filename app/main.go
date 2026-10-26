package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading Message: ", err.Error())
			os.Exit(1)
		}

		request, _, err := decodeRESP(buf[:n])
		if err != nil {
			fmt.Println("Error decoding request: ", err.Error())
			os.Exit(1)
		}

		response := handleCommand(request)
		encodedResponse, err := encodeRESP(response)
		if err != nil {
			fmt.Println("Error encoding response: ", err.Error())
			os.Exit(1)
		}

		_, err = conn.Write(encodedResponse)
		if err != nil {
			fmt.Println("Error writing message: ", err.Error())
			os.Exit(1)
		}
	}
}
