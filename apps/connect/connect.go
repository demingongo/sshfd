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
	Config       map[string]string
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

	host := HostConfig{}

	for _, s := range configLines {
		line := r.FindString(s)
		if line != "" {
			if host.Host != "" {
				result[host.Host] = host
			}
			host = HostConfig{
				Host:   strings.Trim(rHost.ReplaceAllString(line, ""), " "),
				Config: make(map[string]string),
			}
		} else if host.Host != "" {
			trimmedLine := strings.Trim(s, " ")

			if trimmedLine == "Host *" {
				if host.Host != "" {
					result[host.Host] = host
				}
				host = HostConfig{}
			} else if trimmedLine != "" {
				param := strings.Split(trimmedLine, " ")
				key := strings.Trim(param[0], " ")
				value := strings.Trim(param[1], " ")

				if key == "Hostname" {
					host.Hostname = value
				} else if key == "Port" {
					host.Port = value
				} else if key == "User" {
					host.User = value
				} else if key == "IdentityFile" {
					host.IdentityFile = value
				} else {
					host.Config[key] = value
				}
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
			homeDir, err := os.UserHomeDir()
			if err != nil {
				logger.Fatalf("UserHomeDir: %v", err)
			}
			privateKeyPath := strings.ReplaceAll(val.IdentityFile, "~", homeDir)
			privateKey, err := os.ReadFile(privateKeyPath)
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
