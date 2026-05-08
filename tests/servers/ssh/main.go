package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

var port = getEnv("MOCK_SSH_PORT", "18082")

func main() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Accept any password for testing
			return nil, nil
		},
	}

	// Generate host key
	hostKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate host key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Mock SSH server starting on :%s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleSSHConnection(conn, config)
	}
}

func handleSSHConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("SSH handshake failed: %v", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)
	handleChannels(chans)
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "only session channels supported")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Channel accept error: %v", err)
			continue
		}

		go handleSession(channel, requests)
	}
}

func handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			handleExec(channel, req)
		case "shell":
			handleShell(channel)
		case "pty-req":
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func handleExec(channel ssh.Channel, req *ssh.Request) {
	var execReq struct {
		Command string
	}
	if err := ssh.Unmarshal(req.Payload, &execReq); err != nil {
		req.Reply(false, nil)
		return
	}

	req.Reply(true, nil)

	cmd := execReq.Command
	log.Printf("Executing: %s", cmd)

	output := executeMockCommand(cmd)
	channel.Write([]byte(output))
	channel.SendRequest("exit-status", false, ssh.Marshal(&struct{ ExitStatus uint32 }{ExitStatus: 0}))
}

func handleShell(channel ssh.Channel) {
	channel.Write([]byte("Mock SSH shell. Type commands:\n"))
	buf := make([]byte, 1024)
	for {
		n, err := channel.Read(buf)
		if err != nil {
			break
		}
		cmd := strings.TrimSpace(string(buf[:n]))
		if cmd == "exit" {
			break
		}
		output := executeMockCommand(cmd)
		if _, err := channel.Write([]byte(output + "\n")); err != nil {
			fmt.Println(err)
		}
	}
}

func executeMockCommand(cmd string) string {
	// Handle failure signal
	if strings.Contains(cmd, "fail") {
		return fmt.Sprintf("ERROR: command failed: %s", cmd)
	}

	// Handle echo
	if strings.HasPrefix(cmd, "echo ") {
		return strings.TrimPrefix(cmd, "echo ")
	}

	// Handle hostname
	if cmd == "hostname" {
		host, _ := os.Hostname()
		return host
	}

	// Handle whoami
	if cmd == "whoami" {
		return "testuser"
	}

	// Handle ls
	if cmd == "ls" || strings.HasPrefix(cmd, "ls ") {
		return "file1.txt\nfile2.txt\nconfig.yaml"
	}

	// Handle pwd
	if cmd == "pwd" {
		return "/home/testuser"
	}

	// Handle date
	if cmd == "date" {
		return "Mon Jan 01 00:00:00 UTC 2026"
	}

	// Default
	return fmt.Sprintf("mock output for: %s", cmd)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// Ensure io is used
var _ io.Reader
