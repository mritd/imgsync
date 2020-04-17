package cmd

import (
	"github.com/mritd/gcrsync/gcrsync"
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test sync",
	Long: `
Test sync.`,
	Run: func(cmd *cobra.Command, args []string) {

		gcr := &gcrsync.Gcr{
			Proxy:        proxy,
			NameSpace:    nameSpace,
			QueryLimit:   queryLimit,
			ProcessLimit: processLimit,
			SyncTimeOut:  syncTimeOut,
			HttpTimeOut:  httpTimeOut,
			TestMode:     true,
		}
		gcr.Init()
		gcr.Sync()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
