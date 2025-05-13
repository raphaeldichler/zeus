package main

import (
	"fmt"

	"github.com/raphaeldichler/zeus/internal/ingress"
)

func main() {
  fmt.Println(ingress.NginxConfDefault)
  /*
	// Define tmpfs file location (in-memory)
	tmpfsDir := "/dev/shm"
	fileName := "secret.txt"
	fullPath := filepath.Join(tmpfsDir, fileName)

	// The secret to store
	secretData := "API_KEY=super-secret-value\nDB_PASSWORD=very-secret-password"

	// Create and write the file
	err := os.WriteFile(fullPath, []byte(secretData), 0600)
	if err != nil {
		fmt.Printf("Failed to write secret file: %v\n", err)
		return
	}

	fmt.Printf("Secret file written to: %s\n", fullPath)
	fmt.Println("To copy into your Docker container, run the following command:")
	fmt.Printf("  docker cp %s <container_id>:/path/in/container/%s\n", fullPath, fileName)
  */
}

