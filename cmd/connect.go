/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/demingongo/sshfd/apps/connect"
	"github.com/demingongo/sshfd/globals"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a remote host",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals.LoadGlobals()
		connect.Run()
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	connectCmd.PersistentFlags().String("host", "", "host")

	viper.BindPFlag("host", connectCmd.PersistentFlags().Lookup("host"))
}
