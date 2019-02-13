package cmd

import (
	"github.com/mritd/gcrsync/gcrsync"
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor sync images",
	Long: `
Monitor sync images.`,
	Run: func(cmd *cobra.Command, args []string) {

		gcr := &gcrsync.Gcr{
			Proxy:          proxy,
			DockerUser:     dockerUser,
			DockerPassword: dockerPassword,
			NameSpace:      nameSpace,
			QueryLimit:     make(chan int, queryLimit),
			ProcessLimit:   make(chan int, processLimit),
			SyncTimeOut:    syncTimeOut,
			HttpTimeOut:    httpTimeOut,
			GithubRepo:     githubRepo,
			GithubToken:    githubToken,
			CommitMsg:      commitMsg,
			MonitorCount:   monitorCount,
			Debug:          debug,
		}
		gcr.Init()
		gcr.Monitor()
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.PersistentFlags().IntVar(&monitorCount, "count", -1, "monitor count")
}
