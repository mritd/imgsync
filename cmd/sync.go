package cmd

import (
	"github.com/mritd/gcrsync/gcrsync"
	"github.com/spf13/cobra"
)

var commitMsg string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync gcr images",
	Long: `
Sync gcr images.`,
	Run: func(cmd *cobra.Command, args []string) {

		gcr := &gcrsync.Gcr{
			Proxy:        proxy,
			NameSpace:    nameSpace,
			QueryLimit:   queryLimit,
			ProcessLimit: processLimit,
			SyncTimeOut:  syncTimeOut,
			HttpTimeOut:  httpTimeOut,
		}
		gcr.Init()
		gcr.Sync()
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVar(&commitMsg, "commitmsg", "Travis CI Auto Synchronized.", "commit message")
}
