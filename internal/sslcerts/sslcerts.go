package sslcerts

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mjefferson-whs/listener/internal/jsonlog"
)

// installMkcert downloads and installs mkcert if not found
func installMkcert(downloadURL, installPath string, logger *jsonlog.Logger) error {
	logger.PrintInfo("mkcert not found, installing...", nil)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download mkcert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response downloading mkcert: %s", resp.Status)
	}

	out, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save mkcert: %w", err)
	}

	// Make executable on Linux/macOS
	if runtime.GOOS != "windows" {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make mkcert executable: %w", err)
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
func GenerateSSLCert(tlsDirPath, ip string, logger *jsonlog.Logger) error {
	var downloadURL, mkcertPath string
	// TODO: Add architecture checks
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-windows-amd64.exe"
		mkcertPath = filepath.Join(os.Getenv("LocalAppData"), "Programs", "kamar-listener", "mkcert.exe")
	case "linux":
		downloadURL = "https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64"
		mkcertPath = "/usr/local/bin/mkcert"
	case "darwin":
		downloadURL = "https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-darwin-arm64"
		mkcertPath = "/usr/local/bin/mkcert"
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Check if mkcert is installed
	if _, err := exec.LookPath("mkcert"); err != nil {
		if err := installMkcert(downloadURL, mkcertPath, logger); err != nil {
			return err
		}
	} else {
		mkcertPath = "mkcert"
	}

	// Install mkcert root CA
	if !isRootCAInstalled() {
		logger.PrintInfo("Running mkcert -install...", nil)
		cmd := exec.Command(mkcertPath, "-install")
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
	cmd := exec.Command(mkcertPath, "-key-file", keyPath, "-cert-file", certPath, "localhost", ip)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	logger.PrintInfo("SSL certificate generated successfully: key.pem & cert.pem", nil)
	return nil
}
