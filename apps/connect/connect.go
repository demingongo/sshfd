package connect

import (
	"fmt"
	"os"

	//"os/exec"
	"regexp"
	"sshfd/globals"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

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

func Run() {
	logger := globals.Logger
	host := ""

	if viper.GetString("host") != "" {
		host = viper.GetString("host")
	} else {
		// select a host

		configBody := strings.Split(strings.ReplaceAll(loadConfig(), "\r\n", "\n"), "\n")

		r, err := regexp.Compile("Host ([^*]+)")

		if err != nil {
			logger.Fatalf("%v", err)
		}

		r2, err := regexp.Compile("Host ")
		if err != nil {
			logger.Fatalf("%v", err)
		}

		var hosts []string

		for _, s := range configBody {
			//fmt.Println(i, s)
			line := r.FindString(s)
			if line != "" {
				hosts = append(hosts, r2.ReplaceAllString(line, ""))
			}
		}

		form := runFormSelectHost(
			"",
			hosts,
		)
		if form.State == huh.StateCompleted {
			if host = form.Get("host").(string); host != "" {
				logger.Debug(fmt.Sprintf("selected \"%s\"", host))
			}
		}

		for _, s := range hosts {
			logger.Debug(fmt.Sprint(s))
		}
	}

	if host != "" {
		/*
			logger.Debug(fmt.Sprintf("connect to \"%s\"", host))
			cmd := exec.Command("ssh", host)
			stdout, err := cmd.Output()
			if err != nil {
				logger.Printf("Could not connect to \"%s\"", host)
				logger.Debugf("%v", stdout)
				logger.Fatalf("%v", err)
			}
			logger.Debug(fmt.Sprint(string(stdout)))
		*/
		logger.Printf("ssh %s", host)
	} else {
		logger.Fatal("No host")
	}
}
