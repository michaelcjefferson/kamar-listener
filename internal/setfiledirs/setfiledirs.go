package setfiledirs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type ApplicationDataDirPaths struct {
	ApplicationDir string
	FileDirs       map[string]string
}

// Get path to currently logged in user's directory with open write permissions, create a "kamar-listener" folder, and return that as the application data path. On Windows: \%USERPROFILE%\Documents\kamar-listener, on iOS: $HOME/Applications/kamar-listener, on Linux: $HOME/.local/share/kamar-listener
func SetFileDirs(dirName string, fileDirNames []string) (*ApplicationDataDirPaths, error) {
	var dirPaths ApplicationDataDirPaths
	dirPaths.FileDirs = make(map[string]string)

	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to locate the user home directory: %v", err)
	}

	// Determine the required directory based on OS
	var requiredDir string
	switch runtime.GOOS {
	case "windows":
		requiredDir = filepath.Join(os.Getenv("LocalAppData"), "Programs", dirName)
	case "darwin":
		requiredDir = filepath.Join(userHome, "Applications", dirName)
	case "linux":
		requiredDir = filepath.Join(userHome, ".local", "share", dirName)
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	dirPaths.ApplicationDir = requiredDir

	// Make dir with write permissions for the owner, and read and exec permissions for all others in group
	if err := os.MkdirAll(requiredDir, 0755); err != nil {
		return &dirPaths, fmt.Errorf("failed to create application data directory: %w", err)
	}

	// Create file dirs inside required dir
	for _, name := range fileDirNames {
		fileDir := filepath.Join(requiredDir, name)
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return &dirPaths, fmt.Errorf("failed to create %s data directory: %w", name, err)
		}
		dirPaths.FileDirs[name] = fileDir
	}

	return &dirPaths, nil
}
