package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func main() {
	socketPath := "/run/zeus/zeusd.sock"

	if err := os.RemoveAll(socketPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to remove existing socket: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll("/run/zeus", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create /run/zeus: %v\n", err)
		os.Exit(1)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on socket: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Optional: restrict permissions of the socket file
	if err := os.Chmod(socketPath, 0660); err != nil {
		fmt.Fprintf(os.Stderr, "failed to chmod socket: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, world! YEA YEAH")
	})

	fmt.Println("Server listening on", socketPath)
	if err := http.Serve(listener, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
	}
}
