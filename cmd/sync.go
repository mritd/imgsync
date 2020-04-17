package cmd

import (
	"time"

	"github.com/mritd/gcrsync/gcrsync"
	"github.com/spf13/cobra"
)

var gcr gcrsync.Gcr

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync gcr images",
	Long: `
Sync gcr images.`,
	Run: func(cmd *cobra.Command, args []string) {
		gcr.Init().Sync()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVar(&gcr.DockerHubUser, "user", "", "docker registry user")
	syncCmd.PersistentFlags().StringVar(&gcr.DockerHubPassword, "password", "", "docker registry user password")
	syncCmd.PersistentFlags().StringVar(&gcr.NameSpace, "namespace", "google-containers", "google container registry namespace")
	syncCmd.PersistentFlags().IntVar(&gcr.QueryLimit, "querylimit", 50, "http query limit")
	syncCmd.PersistentFlags().IntVar(&gcr.ProcessLimit, "processlimit", 10, "image process limit")
	syncCmd.PersistentFlags().DurationVar(&gcr.HttpTimeOut, "httptimeout", 10*time.Second, "http request timeout")
	syncCmd.PersistentFlags().DurationVar(&gcr.SyncTimeOut, "synctimeout", 1*time.Hour, "sync timeout")
}
