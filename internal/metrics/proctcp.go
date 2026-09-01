package metrics

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/iammuuo/notrust/internal/docker"
)

const tcpStateEstablished = "01"

// hasEstablishedConnections reports whether any of the container's
// published host ports currently has an ESTABLISHED connection.
func hasEstablishedConnections(ports []docker.PortBinding) (bool, error) {
	hostPorts := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		if p.HostPort != 0 {
			hostPorts[p.HostPort] = struct{}{}
		}
	}
	if len(hostPorts) == 0 {
		return false, nil
	}

	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		found, err := scanEstablished(path, hostPorts)
		if err != nil {
			if os.IsNotExist(err) {
				continue // tcp6 absent on some hosts, not fatal
			}
			return false, err
		}
		if found {
			return true, nil
		}
	}
	return false, nil
}

func scanEstablished(path string, hostPorts map[int]struct{}) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // discard header line
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || fields[3] != tcpStateEstablished {
			continue
		}
		if _, port, err := parseLocalAddress(fields[1]); err == nil {
			if _, ok := hostPorts[port]; ok {
				return true, nil
			}
		}
	}
	return false, scanner.Err()
}

// parseLocalAddress decodes /proc/net/tcp's "IP:PORT" hex field,
// e.g. "0100007F:1F90". We only need the port, so the IP is discarded.
func parseLocalAddress(field string) (string, int, error) {
	parts := strings.Split(field, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("malformed local_address %q", field)
	}
	portBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(portBytes) != 2 {
		return "", 0, fmt.Errorf("malformed port in %q", field)
	}
	return parts[0], int(portBytes[0])<<8 | int(portBytes[1]), nil
}
