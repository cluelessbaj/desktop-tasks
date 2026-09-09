package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func getSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/tmp"
	}
	return filepath.Join(runtimeDir, "desktop-tasks.sock")
}

func SendIPCCommand(command string) bool {
	sockPath := getSocketPath()
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return false
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\n", strings.TrimSpace(command))
	return err == nil
}

func startIPCServer(sockPath string, onCommand func(cmd string)) error {
	_ = os.Remove(sockPath)

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				if scanner.Scan() {
					cmd := strings.TrimSpace(scanner.Text())
					if cmd != "" {
						onCommand(cmd)
					}
				}
			}(conn)
		}
	}()

	return nil
}
