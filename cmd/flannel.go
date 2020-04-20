package cmd

import (
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
	flannelCmd.PersistentFlags().StringVar(&flannel.DockerUser, "user", "", "docker hub user")
	flannelCmd.PersistentFlags().StringVar(&flannel.DockerPassword, "password", "", "docker hub user password")
	flannelCmd.PersistentFlags().StringVar(&flannel.Proxy, "proxy", "", "http client proxy")
	flannelCmd.PersistentFlags().StringVar(&flannel.IgnoreTagRex, "ignoretag", "", "ignore image where tag matches regular expression")
	flannelCmd.PersistentFlags().IntVar(&flannel.ProcessLimit, "processlimit", core.DefaultLimit, "image process limit")
	flannelCmd.PersistentFlags().DurationVar(&flannel.HTTPTimeOut, "httptimeout", core.DefaultHTTPTimeOut, "http request timeout")
	flannelCmd.PersistentFlags().DurationVar(&flannel.SyncTimeOut, "synctimeout", core.DefaultSyncTimeout, "docker hub sync timeout")
}
