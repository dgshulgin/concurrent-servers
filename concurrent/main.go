package main

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"
)

const (
	MaxBufferSize = 32
)

func main() {

	service := "localhost:1200"
	err := run(service)
	if err != nil {
		log.Fatalf("Fatal: %s\n", err.Error())
	}
}

func listen(service string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		return nil, err
	}

	return net.ListenTCP("tcp", tcpAddr)
}

func handshake(conn net.Conn) (string, error) {
	_, err := conn.Write([]byte("*"))
	if err != nil {
		return "", err
	}

	buf := make([]byte, MaxBufferSize)
	_, err = conn.Read(buf)
	if err != nil {
		return "", err
	}

	name := strings.Trim(string(buf), "\r\n")

	return string(name), nil
}

func run(service string) error {

	listener, err := listen(service)
	if err != nil {
		return errors.Join(errors.New("listen failed"), err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Join(errors.New("accept failed"), err)
		}

		go func() {
			log.Printf("Connection: %s\n", conn.RemoteAddr().String())

			name, err := handshake(conn)
			if err != nil || len(name) == 0 {
				log.Printf("handshake failed: %s", err.Error())
				return //errors.Join(errors.New("handshake failed"), err)
			}
			log.Printf("Processing client: %s\n", name)

			err = serveConnection(conn, name)
			if err != nil {
				log.Printf("service failed: %s", err.Error())
				return //errors.Join(errors.New("service failed"), err)
			}

			log.Printf("Connection %s closed\n", name)
		}()
	}
}

type ProcessingState int

const (
	WaitForMsg ProcessingState = iota // состояние сервера: ожидаю сообщения
	InMsg                             // состояние сервера: обрабатываю сообщение
)

// Протокол:
// - сервер посылает * для потверждения соединения
// - клиент возвращает идентификатор;
// - клиентское сообщение находится между знаками ^ (начало сообщения) и $ (конец сообщения)
// - содержимое клиентского сообщения за пределами пары ^$ игнорируется
// - в ответ на клиентское сообщение сервер высылает ответ в виде byte+1 на каждый символ клиентского сообщения

func serveConnection(conn net.Conn, client string) error {

	state := WaitForMsg

	for {
		buf := make([]byte, MaxBufferSize)
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return errors.Join(errors.New("socket read failed"), err)
		}

		if n == 0 {
			continue
		}

		buf = append([]byte(nil), buf[:n]...)

		var out []byte
		for _, bb := range buf {
			switch state {
			case WaitForMsg:
				if bb == '^' {
					state = InMsg
				}
			case InMsg:
				if bb == '$' {
					state = WaitForMsg
					continue
				}
				out = append(out, bb+1)
			}
		}
		if len(out) == 0 {
			out = []byte("skipped")
		}

		log.Printf("client %s msg: %s -> %s\n", client, string(buf), string(out))
		_, err = conn.Write(out)
		if err != nil {
			return nil
		}
	}
}
