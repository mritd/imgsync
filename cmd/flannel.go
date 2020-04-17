package cmd

import (
	"time"

	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var flannel core.Flannel

var flannelCmd = &cobra.Command{
	Use:   "flannel",
	Short: "Sync flannel images",
	Long: `
Sync flannel images.`,
	Run: func(cmd *cobra.Command, args []string) {
		flannel.Init().Sync()
	},
}

func init() {
	rootCmd.AddCommand(flannelCmd)
	flannelCmd.PersistentFlags().StringVar(&flannel.DockerHubUser, "user", "", "docker hub user")
	flannelCmd.PersistentFlags().StringVar(&flannel.DockerHubPassword, "password", "", "docker hub user password")
	flannelCmd.PersistentFlags().StringVar(&flannel.Proxy, "proxy", "", "http client proxy")
	flannelCmd.PersistentFlags().IntVar(&flannel.ProcessLimit, "processlimit", 10, "image process limit")
	flannelCmd.PersistentFlags().DurationVar(&flannel.HttpTimeOut, "httptimeout", 10*time.Second, "http request timeout")
	flannelCmd.PersistentFlags().DurationVar(&flannel.SyncTimeOut, "synctimeout", 1*time.Hour, "docker hub sync timeout")
}
