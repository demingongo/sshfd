/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/demingongo/sshfd/apps/download"
	"github.com/demingongo/sshfd/globals"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a remote file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals.LoadGlobals()
		download.Run()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	// Point to a string variable in which to store the value of the flag (needed for same name flags between commands)
	downloadCmd.PersistentFlags().StringVarP(&globals.LocalFileFlag, "local-file", "l", "", "local file path (target)")
	downloadCmd.PersistentFlags().StringVarP(&globals.RemoteFileFlag, "remote-file", "r", "", "remote file (source)")

	downloadCmd.MarkPersistentFlagRequired("local-file")
	downloadCmd.MarkPersistentFlagRequired("remote-file")

	downloadCmd.MarkPersistentFlagFilename("local-file")

	viper.BindPFlag("localFile", downloadCmd.PersistentFlags().Lookup("local-file"))
	viper.BindPFlag("remoteFile", downloadCmd.PersistentFlags().Lookup("remote-file"))
}
