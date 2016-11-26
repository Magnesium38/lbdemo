package common

import "net"

// GetOpenPort returns an available port as an integer.
func GetOpenPort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port
	return port, nil
}
