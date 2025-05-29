package stat

import (
	"os"

	"sshfd/globals"
	"sshfd/utils"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func Run() {
	logger := globals.Logger

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {

		client, err := utils.DialSsh(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer session.Close()

		modes := ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}

		if err := session.RequestPty("linux", 80, 40, modes); err != nil {
			logger.Fatal("Request for pseudo terminal failed:", err)
		}

		session.Stdout = os.Stdout
		session.Stdin = os.Stdin
		session.Stderr = os.Stderr

		if err := session.Shell(); err != nil {
			logger.Fatal("Failed to start shell:", err)
		}

		if err := session.Wait(); err != nil {
			logger.Fatal("Failed to run:", err)
		}

	} else {
		logger.Fatal("No host")
	}
}
