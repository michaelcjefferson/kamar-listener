package checkrundir

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Ensure that binary is being run from the required directory - \User\Documents\listener on Windows, /user/Applications\listener on iOS, ~/.local/share/listener on Linux
func EnforceRunLocation() {
	// Get the directory the binary is currently running in
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get information about the directory this app is being run from: %v\nThis application needs to be run from a folder called 'listener' inside your Documents folder.", err)
	}

	execDir := filepath.Dir(execPath)
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to locate the user home directory: %v", err)
	}

	// Determine the required directory based on OS
	var requiredDir string
	switch runtime.GOOS {
	case "windows":
		requiredDir = filepath.Join(userHome, "Documents", "kamar-listener")
	case "darwin":
		requiredDir = filepath.Join(userHome, "Applications", "kamar-listener")
	case "linux":
		requiredDir = filepath.Join(userHome, ".local", "share", "kamar-listener")
	default:
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}

	// If run directory doesn't match expected directory, terminate the program
	if filepath.Clean(execDir) != filepath.Clean(requiredDir) {
		log.Fatalf("This application must be run from the following directory:\n%s\nCurrent directory:\n%s\n", requiredDir, execDir)
	}
}
