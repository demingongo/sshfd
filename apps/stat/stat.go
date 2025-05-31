package stat

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/demingongo/sshfd/globals"
	"github.com/demingongo/sshfd/utils"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type DiskStat struct {
	Filesystem string
	Type       string
	Size       string
	Used       string
	Available  string
	UsePercent string
	MountedOn  string
}

func Run() {
	logger := globals.Logger

	if val, ok := utils.LoadHostConfig(viper.GetString("host")); ok && val.Hostname != "" {

		client, err := utils.DialSsh(val)
		if err != nil {
			logger.Fatalf("Unable to connect: %v", err)
		}
		defer client.Close()

		session, err := createSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer session.Close()

		var b bytes.Buffer
		session.Stdout = &b // get output

		if err := session.Run("df -Th"); err != nil {
			logger.Error(b.String())
			logger.Fatal("Failed to run:", err)
		}

		logger.Debug(b.String())

		// get the lines and remove the first line ([1:]) as it is the columns header
		dfLines := strings.Split(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n")[1:]

		var disksStats []DiskStat

		for _, line := range dfLines {
			cols := filter(
				strings.Split(strings.Trim(line, ""), " "),
				isNotEmpty,
			)

			/*
			* 0 = Filesystem
			* 1 = Type
			* 2 = Size
			* 3 = Used
			* 4 = Available
			* 5 = Use%
			* 6 = Mounted on
			 */

			if len(cols) < 7 {
				continue
			}

			if cols[1] == "tmpfs" || cols[1] == "devtmpfs" || cols[1] == "efivarfs" {
				continue
			}

			disksStats = append(disksStats, DiskStat{
				Filesystem: cols[0],
				Type:       cols[1],
				Size:       cols[2],
				Used:       cols[3],
				Available:  cols[4],
				UsePercent: cols[5],
				MountedOn:  cols[6],
			})
		}

		logger.Info(fmt.Sprintf("%v", disksStats))

		mSession, err := createSession(client)
		if err != nil {
			logger.Fatalf("Failed to create a session: %v", err)
		}
		defer mSession.Close()

		b.Reset() // empty buffer
		mSession.Stdout = &b

		if err := mSession.Run("free -mh"); err != nil {
			logger.Error(b.String())
			logger.Fatal("Failed to run:", err)
		}

		logger.Info(b.String())

		// get the lines and remove the first line ([1:]) as it is the columns header
		dfLines = strings.Split(strings.ReplaceAll(b.String(), "\r\n", "\n"), "\n")[1:]

		for _, line := range dfLines {
			cols := filter(
				strings.Split(strings.Trim(line, ""), " "),
				isNotEmpty,
			)

			if len(cols) < 4 {
				continue
			}

			if cols[2] == "total" {
				continue
			}

			/*
			* 0 = Type
			* 1 = Total
			* 2 = Used
			* 3 = Free
			* 4 = Shared
			* 5 = Cache
			* 6 = Available
			 */

			logger.Infof("%v", cols)
		}

	} else {
		logger.Fatal("No host")
	}
}

func isNotEmpty(s string) bool { return s != "" }

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func createSession(client *ssh.Client) (*ssh.Session, error) {
	logger := globals.Logger

	session, err := client.NewSession()
	if err != nil {
		logger.Errorf("Failed to create a session: %v", err)
		return session, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("linux", 80, 40, modes); err != nil {
		logger.Errorf("Request for pseudo terminal failed: %v", err)
		return session, err
	}

	return session, err
}
