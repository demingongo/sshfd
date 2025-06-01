package download

import (
	"context"
	"fmt"
	"os"

	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"
	"github.com/spf13/viper"
)

func Run() {
	logger := globals.Logger

	localFile := viper.GetString("localFile")
	remoteFile := viper.GetString("remoteFile")

	if localFile == "" {
		logger.Fatal("Missing local file path (target)")
	}

	if remoteFile == "" {
		logger.Fatal("Missing remote file (source)")
	}

	logger.Debug(fmt.Sprintf("localFile %s", localFile))
	logger.Debug(fmt.Sprintf("remoteFile %s", remoteFile))

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {
		client, err := utils.CreateSCPClient(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		// Open the source file
		f, _ := os.Create(localFile)

		// Close the file after it has been copied
		defer f.Close()

		// Finally, copy the file over
		// Usage: CopyFromFile(context, file, remotePath, permission)

		// the context can be adjusted to provide time-outs or inherit from other contexts if this is embedded in a larger application.
		err = client.CopyFromRemote(context.Background(), f, remoteFile)

		if err != nil {
			logger.Fatalf("Error while copying file: %v", err)
		}

		logger.Info(fmt.Sprintf("File downloaded to %s", localFile))

	} else {
		logger.Fatal("No host")
	}
}
