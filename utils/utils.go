package utils

import (
	"fmt"
	"os"
	"slices"

	"regexp"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/demingongo/sshfd/globals"

	"github.com/charmbracelet/huh"
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

func LoadHostConfig(host string) (HostConfig, bool) {
	logger := globals.Logger

	hosts := loadHostConfigs()

	//logger.Debugf("hosts: %v", hosts)

	if host == "" {
		// select a host

		var keys []string

		for k := range hosts {
			keys = append(keys, k)
		}

		slices.Sort(keys)

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

	val, ok := hosts[host]

	return val, ok
}

func createClientConfig(hc HostConfig) *ssh.ClientConfig {
	logger := globals.Logger

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if hc.User != "" {
		config.User = hc.User
	}

	if hc.IdentityFile != "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Fatalf("UserHomeDir: %v", err)
		}
		privateKeyPath := strings.ReplaceAll(hc.IdentityFile, "~", homeDir)
		privateKey, err := os.ReadFile(privateKeyPath)
		if err != nil {
			logger.Fatalf("Unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			errString := fmt.Sprintf("%v", err)
			logger.Debugf("error string: %s", errString)
			if strings.Contains(errString, "passphrase protected") {
				var passphrase string
				form := runFormPassphrase(
					"",
				)
				if form.State == huh.StateCompleted {
					passphrase = form.Get("passphrase").(string)
				}
				if passphrase == "" {
					logger.Debug("Empty passphrase")
					logger.Fatalf("Unable to parse private key: %v", err)
				}
				signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
				logger.Debug("ParsePrivateKeyWithPassphrase")
				if err != nil {
					logger.Fatalf("Unable to parse private key: %v", err)
				}
			} else {
				logger.Fatalf("Unable to parse private key: %v", err)
			}
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	return config
}

func DialSsh(hc HostConfig) (*ssh.Client, error) {
	logger := globals.Logger

	logger.Debugf("DialSsh %v", hc)

	config := createClientConfig(hc)

	hostname := hc.Hostname + ":"

	if hc.Port != "" {
		hostname += hc.Port
	} else {
		hostname += "22"
	}

	client, err := ssh.Dial("tcp", hostname, config)

	return client, err
}

func CreateSession(client *ssh.Client) (*ssh.Session, error) {
	logger := globals.Logger

	session, err := client.NewSession()
	if err != nil {
		logger.Errorf("Failed to create a session: %v", err)
		return session, err
	}

	return session, err
}

func RequestPty(session *ssh.Session) error {
	logger := globals.Logger

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("linux", 40, 80, modes); err != nil {
		logger.Errorf("Request for pseudo terminal failed: %v", err)
		return err
	}

	return nil
}

func CreateSCPClient(hc HostConfig) (*scp.Client, error) {
	logger := globals.Logger

	logger.Debugf("CreateSCPClient %v", hc)

	config := createClientConfig(hc)

	hostname := hc.Hostname + ":"

	if hc.Port != "" {
		hostname += hc.Port
	} else {
		hostname += "22"
	}

	// Create a new SCP client
	client := scp.NewClient(hostname, config)

	// Connect to the remote server
	err := client.Connect()
	if err != nil {
		//fmt.Println("Couldn't establish a connection to the remote server ", err)
		return &client, err
	}

	return &client, err
}
