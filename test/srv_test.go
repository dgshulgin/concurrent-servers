package main

import (
	"net"
	"testing"
	"time"
)

const (
	MaxBufferSize = 32
)

var messages = []string{
	"^abc$de^abte$f",
	"xyz^123",
	"25$^ab$abab",
	"abc",
	"$abc",
}

var names = []string{
	"Jean", "Paul", "Gautier",
}

var service string = "localhost:1200"

func TestClient(t *testing.T) {

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn, err := connect(service)
			if err != nil {
				t.Fatalf("connect: %s", err.Error())
			}
			defer conn.Close()

			//protocol: read *
			prompt := make([]byte, 1)
			_, err = conn.Read(prompt)
			if err != nil {
				t.Fatalf("violation of protocol STAR: %s", err.Error())
			}
			t.Logf("%s\n", string(prompt))

			// protocol: send id
			_, err = conn.Write([]byte(name))
			if err != nil {
				t.Fatalf("violation of protocol NAME: %s", err.Error())
			}
			t.Logf("%s\n", name)

			// protocol: send messages
			for _, message := range messages {
				out, err := sendRcv(message, conn)
				if err != nil {
					t.Errorf("client %s failed on msg \"%s\" with err: %s", name, message, err.Error())
				} else {
					t.Logf("%s -> %s\n", message, out)
				}
			}
		})
	}

}

func connect(service string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, tcpAddr)
}

func sendRcv(msg string, conn *net.TCPConn) (string, error) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return "", err
	}

	// especially to highlight execution time when parallel running
	time.Sleep(1000 * time.Millisecond)

	buf := make([]byte, MaxBufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	buf = append([]byte(nil), buf[:n]...)

	return string(buf), nil
}
