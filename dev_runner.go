package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Find home directory for air path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home dir: %v\n", err)
		os.Exit(1)
	}
	airPath := filepath.Join(homeDir, "go", "bin", "air")

	// Context for backend
	backendCmd := exec.Command(airPath)
	backendCmd.Stdout = os.Stdout
	backendCmd.Stderr = os.Stderr
	backendCmd.Env = os.Environ() // Pass current environment

	// Context for frontend
	frontendCmd := exec.Command("npm", "run", "dev", "--", "--port", "3500", "-H", "0.0.0.0")
	frontendCmd.Dir = "./frontend"
	frontendCmd.Stdout = os.Stdout
	frontendCmd.Stderr = os.Stderr
	frontendCmd.Env = os.Environ()

	// Start Backend
	fmt.Println("Starting Backend (Air)...")
	if err := backendCmd.Start(); err != nil {
		fmt.Printf("Failed to start backend: %v\n", err)
		os.Exit(1)
	}

	// Start Frontend
	fmt.Println("Starting Frontend (Next.js)...")
	if err := frontendCmd.Start(); err != nil {
		fmt.Printf("Failed to start frontend: %v\n", err)
		// Try to kill backend if frontend fails
		backendCmd.Process.Kill()
		os.Exit(1)
	}

	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal
	<-sigs
	fmt.Println("\nReceived interrupt, shutting down...")

	// Kill processes
	if err := backendCmd.Process.Signal(syscall.SIGTERM); err != nil {
		backendCmd.Process.Kill()
	}
	if err := frontendCmd.Process.Signal(syscall.SIGTERM); err != nil {
		frontendCmd.Process.Kill()
	}

	// Wait for them to exit
	backendCmd.Wait()
	frontendCmd.Wait()
	fmt.Println("Shutdown complete.")
}
