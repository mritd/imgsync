package cmd

import (
	"context"
	"time"

	"github.com/mritd/imgsync/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flTimeout time.Duration
var flSyncOption core.SyncOption

var flannelCmd = &cobra.Command{
	Use:   "flannel",
	Short: "Sync flannel images",
	Long: `
Sync flannel images.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if err := core.LoadManifests(); err != nil {
			logrus.Fatalf("failed to load manifests: %s", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		s, err := core.New("flannel")
		if err != nil {
			logrus.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), gcrTimeout)
		defer cancel()
		s.Sync(ctx, &flSyncOption)
	},
}

func init() {
	rootCmd.AddCommand(flannelCmd)
	flannelCmd.PersistentFlags().StringVar(&flSyncOption.User, "user", "", "docker hub user")
	flannelCmd.PersistentFlags().StringVar(&flSyncOption.Password, "password", "", "docker hub user password")
	flannelCmd.PersistentFlags().StringVar(&flSyncOption.Proxy, "proxy", "", "http client proxy")
	flannelCmd.PersistentFlags().DurationVar(&flSyncOption.Timeout, "httptimeout", core.DefaultHTTPTimeOut, "http request timeout")
	flannelCmd.PersistentFlags().DurationVar(&flTimeout, "timeout", core.DefaultSyncTimeout, "docker hub sync timeout")
}
