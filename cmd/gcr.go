package cmd

import (
	"context"
	"time"

	"github.com/mritd/imgsync/core"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var gcrTimeout time.Duration
var gcrSyncOption core.SyncOption

var gcrCmd = &cobra.Command{
	Use:   "gcr",
	Short: "Sync gcr images",
	Long: `
Sync gcr images.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := core.LoadManifests(); err != nil {
			logrus.Fatalf("failed to load manifests: %s", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		s, err := core.New("gcr")
		if err != nil {
			logrus.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), gcrTimeout)
		defer cancel()
		s.Sync(ctx, gcrSyncOption)
	},
}

func init() {
	rootCmd.AddCommand(gcrCmd)
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.User, "user", "", "docker hub user")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.Password, "password", "", "docker hub user password")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.NameSpace, "namespace", "google-containers", "google container registry namespace")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.Proxy, "proxy", "", "http client proxy")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.QueryLimit, "querylimit", core.DefaultLimit, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.Limit, "processlimit", core.DefaultLimit, "sync image limit")
	gcrCmd.PersistentFlags().DurationVar(&gcrSyncOption.Timeout, "httptimeout", core.DefaultHTTPTimeOut, "http request timeout")
	gcrCmd.PersistentFlags().BoolVar(&gcrSyncOption.Kubeadm, "kubeadm", false, "sync kubeadm images(ignore namespace, use k8s.gcr.io)")
	gcrCmd.PersistentFlags().DurationVar(&gcrTimeout, "timeout", 1*time.Hour, "docker hub sync timeout")
}
