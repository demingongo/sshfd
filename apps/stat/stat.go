package stat

import (
	"bytes"

	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"

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

		var b bytes.Buffer
		session.Stdout = &b // get output

		if err := session.Run("df -Th"); err != nil {
			logger.Error(b.String())
			logger.Fatal("Failed to run:", err)
		}

		logger.Info(b.String())

	} else {
		logger.Fatal("No host")
	}
}
