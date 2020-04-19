package cmd

import (
	"github.com/mritd/imgsync/core"

	"github.com/spf13/cobra"
)

var manifestsCmd = &cobra.Command{
	Use:   "manifests",
	Short: "Download manifests",
	Long: `
Download manifests.`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(manifestsCmd)

	manifestsCmd.PersistentFlags().StringVar(&gcr.NameSpace, "namespace", "google-containers", "google container registry namespace")
	manifestsCmd.PersistentFlags().StringVar(&gcr.NameSpace, "namespace", "google-containers", "google container registry namespace")
	gcrCmd.PersistentFlags().StringVar(&gcr.Proxy, "proxy", "", "http client proxy")
	gcrCmd.PersistentFlags().StringVar(&gcr.IgnoreTagRex, "ignoretag", "", "ignore image where tag matches regular expression")
	gcrCmd.PersistentFlags().IntVar(&gcr.QueryLimit, "querylimit", core.DefaultLimit, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcr.ProcessLimit, "processlimit", core.DefaultLimit, "image process limit")
	gcrCmd.PersistentFlags().DurationVar(&gcr.HTTPTimeOut, "httptimeout", core.DefaultHTTPTimeOut, "http request timeout")
	gcrCmd.PersistentFlags().DurationVar(&gcr.SyncTimeOut, "synctimeout", core.DefaultSyncTimeout, "docker hub sync timeout")
	gcrCmd.PersistentFlags().BoolVar(&gcr.Kubeadm, "kubeadm", false, "sync kubeadm images(ignore namespace, use k8s.gcr.io)")
}
