package cmd

import (
	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var gcr core.Gcr

var gcrCmd = &cobra.Command{
	Use:   "gcr",
	Short: "Sync gcr images",
	Long: `
Sync gcr images.`,
	Run: func(cmd *cobra.Command, args []string) {
		gcr.Init().Sync()
	},
}

func init() {
	rootCmd.AddCommand(gcrCmd)
	gcrCmd.PersistentFlags().StringVar(&gcr.DockerHubUser, "user", "", "docker hub user")
	gcrCmd.PersistentFlags().StringVar(&gcr.DockerHubPassword, "password", "", "docker hub user password")
	gcrCmd.PersistentFlags().StringVar(&gcr.NameSpace, "namespace", "google-containers", "google container registry namespace")
	gcrCmd.PersistentFlags().StringVar(&gcr.Proxy, "proxy", "", "http client proxy")
	gcrCmd.PersistentFlags().IntVar(&gcr.QueryLimit, "querylimit", core.DefaultLimit, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcr.ProcessLimit, "processlimit", core.DefaultLimit, "image process limit")
	gcrCmd.PersistentFlags().DurationVar(&gcr.HTTPTimeOut, "httptimeout", core.DefaultHTTPTimeOut, "http request timeout")
	gcrCmd.PersistentFlags().DurationVar(&gcr.SyncTimeOut, "synctimeout", core.DefaultSyncTimeout, "docker hub sync timeout")
	gcrCmd.PersistentFlags().BoolVar(&gcr.Kubeadm, "kubeadm", false, "sync kubeadm images(ignore namespace, use k8s.gcr.io)")
}
