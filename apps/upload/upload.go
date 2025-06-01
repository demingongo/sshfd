package upload

import (
	"fmt"

	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"
	"github.com/spf13/viper"
)

func Run() {
	logger := globals.Logger

	source := viper.GetString("source")
	target := viper.GetString("target")

	if source == "" {
		logger.Fatal("Missing source (local file)")
	}

	if target == "" {
		logger.Fatal("Missing target (remote path)")
	}

	logger.Debug(fmt.Sprintf("source %s", source))
	logger.Debug(fmt.Sprintf("target %s", target))

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {
		client, err := utils.CreateSCPClient(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		// TODO: upload file
	} else {
		logger.Fatal("No host")
	}
}
