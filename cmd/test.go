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
			Debug:          debug,
			TestMode:       true,
		}
		gcr.Init()
		gcr.Sync()
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
