package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/mritd/imgsync/core"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var version, buildTime, commit string

var debug bool

var rootCmd = &cobra.Command{
	Use:     "imgsync",
	Short:   "Docker image sync tool",
	Version: version,
	Long: `
Docker image sync tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.SetVersionTemplate(versionTpl())
}

func initLog() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func versionTpl() string {
	var tpl = `%s
Name: imgsync
Version: %s
Arch: %s
BuildTime: %s
CommitID: %s
`
	return fmt.Sprintf(tpl, core.Banner, version, runtime.GOOS+"/"+runtime.GOARCH, buildTime, commit)
}

func prerun(_ *cobra.Command, _ []string) {
	if err := core.LoadManifests(); err != nil {
		logrus.Fatalf("failed to load manifests: %s", err)
	}
}

func boot(name string, opt *core.SyncOption) {
	sigs := make(chan os.Signal)
	ctx, cancel := context.WithCancel(context.Background())
	var cancelOnce sync.Once
	defer cancel()
	go func() {
		for range sigs {
			cancelOnce.Do(func() {
				logrus.Info("Receiving a termination signal, gracefully shutdown!")
				cancel()
			})
			logrus.Info("The goroutines pool has stopped, please wait for the remaining tasks to complete.")
		}
	}()
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	core.NewSynchronizer(name).Sync(ctx, opt)
}
