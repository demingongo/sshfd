package connect

import (
	"fmt"
	"os"

	"regexp"
	"sshfd/globals"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type HostConfig struct {
	Host         string
	Hostname     string
	Port         string
	User         string
	IdentityFile string
	OtherConfig  []string
}

func loadConfig() string {

	dirname, err := os.UserHomeDir()
	if err != nil {
		globals.Logger.Fatalf("%v", err)
	}

	body, err := os.ReadFile(dirname + "/.ssh/config")
	if err != nil {
		globals.Logger.Fatalf("%v", err)
	}

	return string(body)
}

func loadHostConfigs() map[string]HostConfig {
	logger := globals.Logger

	result := make(map[string]HostConfig)

	configLines := strings.Split(strings.ReplaceAll(loadConfig(), "\r\n", "\n"), "\n")

	r, err := regexp.Compile("Host ([^*]+)")

	if err != nil {
		logger.Fatalf("%v", err)
	}

	rHost, err := regexp.Compile("Host ")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	rHostname, err := regexp.Compile("Hostname ([^*]+)")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	rHostname2, err := regexp.Compile("Hostname ")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	rPort, err := regexp.Compile("Port ([^*]+)")
	if err != nil {
		logger.Fatalf("%v", err)
	}
	rPort2, err := regexp.Compile("Port ")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	rUser, err := regexp.Compile("User ([^*]+)")
	if err != nil {
		logger.Fatalf("%v", err)
	}
	rUser2, err := regexp.Compile("User ")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	rIdentityFile, err := regexp.Compile("IdentityFile ([^*]+)")
	if err != nil {
		logger.Fatalf("%v", err)
	}
	rIdentityFile2, err := regexp.Compile("IdentityFile ")
	if err != nil {
		logger.Fatalf("%v", err)
	}

	host := HostConfig{}

	for _, s := range configLines {
		line := r.FindString(s)
		if line != "" {
			if host.Host != "" {
				result[host.Host] = host
			}
			host = HostConfig{}
			host.Host = strings.Trim(rHost.ReplaceAllString(line, ""), " ")
		} else {
			trimmedLine := strings.Trim(s, " ")
			found := false

			line = rHostname.FindString(trimmedLine)
			if line != "" {
				host.Hostname = strings.Trim(rHostname2.ReplaceAllString(line, ""), " ")
				found = true
			}

			if !found {
				line = rPort.FindString(trimmedLine)
				if line != "" {
					host.Port = strings.Trim(rPort2.ReplaceAllString(line, ""), " ")
					found = true
				}
			}

			if !found {
				line = rUser.FindString(trimmedLine)
				if line != "" {
					host.User = strings.Trim(rUser2.ReplaceAllString(line, ""), " ")
					found = true
				}
			}

			if !found {
				line = rIdentityFile.FindString(trimmedLine)
				if line != "" {
					host.IdentityFile = strings.Trim(rIdentityFile2.ReplaceAllString(line, ""), " ")
					found = true
				}
			}

			if !found {
				host.OtherConfig = append(host.OtherConfig, strings.Trim(s, " "))
			}
		}
	}

	if host.Host != "" {
		result[host.Host] = host
	}

	return result
}

func Run() {
	logger := globals.Logger
	host := ""

	hosts := loadHostConfigs()

	//logger.Debugf("hosts: %v", hosts)

	if viper.GetString("host") != "" {
		host = viper.GetString("host")
	} else {
		// select a host

		var keys []string

		for k := range hosts {
			keys = append(keys, k)
		}

		form := runFormSelectHost(
			"",
			keys,
		)
		if form.State == huh.StateCompleted {
			if host = form.Get("host").(string); host != "" {
				logger.Debug(fmt.Sprintf("selected \"%s\"", host))
			}
		}
	}

	if val, ok := hosts[host]; host != "" && ok && val.Hostname != "" {

		logger.Debugf("host %v", val)

		config := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		if val.User != "" {
			config.User = val.User
		}

		if val.IdentityFile != "" {
			privateKey, err := os.ReadFile(val.IdentityFile)
			if err != nil {
				logger.Fatalf("Unable to read private key: %v", err)
			}

			signer, err := ssh.ParsePrivateKey(privateKey)
			if err != nil {
				logger.Fatalf("Unable to parse private key: %v", err)
			}

			config.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		}

		hostname := val.Hostname + ":"

		if val.Port != "" {
			hostname += val.Port
		} else {
			hostname += "22"
		}

		client, err := ssh.Dial("tcp", hostname, config)
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
