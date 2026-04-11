package ssh

import (
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func SSHConnection() error {
	user := os.Getenv("VM_HOST")
	ip := os.Getenv("VM_IP")
	auth := os.Getenv("VM_AUTH")

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            user,
		Auth: []ssh.AuthMethod{
			ssh.Password(auth),
		},
		Timeout: 10 * time.Second,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	return nil
}
