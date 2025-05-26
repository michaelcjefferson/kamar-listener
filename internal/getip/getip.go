package getip

import (
	"net"
)

// Get the IP address of the host device - used in mkcert generation (so that the service is accessible by other devices on the same network) and on dashboard display (so that KAMAR can be correctly configured)
func GetLocalIP() (string, error) {
	// Connect to an external address, doesn't matter if it's reachable
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
