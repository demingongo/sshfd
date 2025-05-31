package connect

import (
	"os"

	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"

	"github.com/spf13/viper"
)

func Run() {
	logger := globals.Logger

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {

		client, err := utils.DialSsh(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		session, err := utils.CreateSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer session.Close()

		if err := utils.RequestPty(session); err != nil {
			logger.Fatalf("Request for pseudo terminal failed: %v", err)
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
