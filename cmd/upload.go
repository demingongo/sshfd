/*
Copyright © 2025 demingongo
*/
package cmd

import (
	"github.com/demingongo/sshfd/apps/upload"
	"github.com/demingongo/sshfd/globals"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a local file to a remote host",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		globals.LoadGlobals()
		upload.Run()
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	// Point to a string variable in which to store the value of the flag (needed for same name flags between commands)
	uploadCmd.PersistentFlags().StringVarP(&globals.LocalFileFlag, "local-file", "l", "", "local file (source)")
	uploadCmd.PersistentFlags().StringVarP(&globals.RemoteFileFlag, "remote-file", "r", "", "remote file path (target)")

	uploadCmd.MarkPersistentFlagRequired("local-file")
	uploadCmd.MarkPersistentFlagRequired("remote")

	uploadCmd.MarkPersistentFlagFilename("local-file")

	viper.BindPFlag("localFile", uploadCmd.PersistentFlags().Lookup("local-file"))
	viper.BindPFlag("remoteFile", uploadCmd.PersistentFlags().Lookup("remote-file"))
}
