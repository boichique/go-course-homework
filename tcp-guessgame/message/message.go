package message

import "net"

const (
	Start        = "guess"
	MinMaxFormat = "[%d %d]"
	Higher       = "higher"
	Lower        = "lower"
	Correct      = "correct"
)

func Read(conn net.Conn) (string, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}

func Write(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}
