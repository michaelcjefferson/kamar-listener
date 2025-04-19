package sslcerts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mjefferson-whs/listener/internal/jsonlog"
)

// TODO: Modify to just check for the existence of both mkcert -CAROOT and key.pem/cert.pem in the correct directory. If this fails, prompt the user to install and run mkcert to generate certificates
// TODO UPDATE: ensure the user is running the app from their Documents folder (or Applications for Mac, /usr/bin for Linux?), and if they are, generate SSL certs OR generate SSL certs and DBs in another standard location so that the app can be run from anywhere
// installMkcert downloads and installs mkcert if not found
func installMkcert(logger *jsonlog.Logger) error {
	logger.PrintInfo("mkcert not found, installing...", nil)

	var downloadURL, installPath string
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-windows-amd64.exe"
		installPath = "C:\\Program Files\\mkcert.exe"
	case "linux":
		downloadURL = "https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-linux-amd64"
		installPath = "/usr/local/bin/mkcert"
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Download mkcert
	cmd := exec.Command("curl", "-Lo", installPath, downloadURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download mkcert: %w", err)
	}

	// Make it executable (Linux only)
	if runtime.GOOS == "linux" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to set mkcert executable: %w", err)
		}
	}

	logger.PrintInfo("mkcert installed successfully", nil)
	return nil
}

// isRootCAInstalled checks if the root CA is already installed
// TODO: Not working
func isRootCAInstalled() bool {
	cmd := exec.Command("mkcert", "-CAROOT")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Check if the CA directory exists
	caRoot := string(output)
	_, err = os.Stat(caRoot)
	return err == nil
}

// generateSSLCert runs mkcert to create a trusted certificate pair in the provided TLS directory path
func GenerateSSLCert(tlsDirPath string, logger *jsonlog.Logger) error {
	// Check if mkcert is installed
	if _, err := exec.LookPath("mkcert"); err != nil {
		if err := installMkcert(logger); err != nil {
			return err
		}
	}

	// Install mkcert root CA
	if !isRootCAInstalled() {
		logger.PrintInfo("Running mkcert -install...", nil)
		cmd := exec.Command("mkcert", "-install")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run mkcert -install: %w", err)
		}
	}

	// Ensure the TLS directory exists
	if err := os.MkdirAll(tlsDirPath, 0755); err != nil {
		return fmt.Errorf("failed to create TLS directory: %w", err)
	}

	// Define certificate paths
	certPath := filepath.Join(tlsDirPath, "cert.pem")
	keyPath := filepath.Join(tlsDirPath, "key.pem")

	// Check if cert already exists
	if _, err := os.Stat(certPath); err == nil {
		logger.PrintInfo("SSL certificate already exists, skipping generation.", nil)
		return nil
	}

	// Generate SSL certificate
	logger.PrintInfo("Generating localhost SSL certificate...", nil)
	cmd := exec.Command("mkcert", "-key-file", keyPath, "-cert-file", certPath, "localhost")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	logger.PrintInfo("SSL certificate generated successfully: key.pem & cert.pem", nil)
	return nil
}
