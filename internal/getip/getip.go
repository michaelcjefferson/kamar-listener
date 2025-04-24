package getip

import (
	"fmt"
	"net"
)

// Get the IP address of the host device - used in mkcert generation (so that the service is accessible by other devices on the same network) and on dashboard display (so that KAMAR can be correctly configured)
func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				ipv4 := ipNet.IP.To4()
				fmt.Printf("ipv4 found: %s\n", ipv4)
				if ipv4 != nil {
					return ipv4.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no non-loopback IP found")
}
