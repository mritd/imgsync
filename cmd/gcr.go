package cmd

import (
	"time"

	"github.com/mritd/imgsync/imgsync"
	"github.com/spf13/cobra"
)

var gcr imgsync.Gcr

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
	gcrCmd.PersistentFlags().IntVar(&gcr.QueryLimit, "querylimit", 50, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcr.ProcessLimit, "processlimit", 10, "image process limit")
	gcrCmd.PersistentFlags().DurationVar(&gcr.HttpTimeOut, "httptimeout", 10*time.Second, "http request timeout")
	gcrCmd.PersistentFlags().DurationVar(&gcr.SyncTimeOut, "synctimeout", 1*time.Hour, "docker hub sync timeout")
}
