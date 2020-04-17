package cmd

import (
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:   "gcrsync",
	Short: "A docker image sync tool for Google container registry (gcr.io)",
	Long: `
A docker image sync tool for Google container registry (gcr.io).`,
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
