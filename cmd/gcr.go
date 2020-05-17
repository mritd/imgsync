package cmd

import (
	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var gcrSyncOption core.SyncOption

var gcrCmd = &cobra.Command{
	Use:   "gcr",
	Short: "Sync gcr images",
	Long: `
Sync gcr images.`,
	PreRun: prerun,
	Run: func(cmd *cobra.Command, args []string) {
		boot("gcr", &gcrSyncOption)
	},
}

func init() {
	rootCmd.AddCommand(gcrCmd)
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.User, "user", "", "docker hub user")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.Password, "password", "", "docker hub user password")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.NameSpace, "namespace", "google-containers", "google container registry namespace")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.QueryLimit, "query-limit", core.DefaultLimit, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.Limit, "process-limit", core.DefaultLimit, "sync image limit")
	gcrCmd.PersistentFlags().DurationVar(&gcrSyncOption.Timeout, "timeout", core.DefaultSyncTimeout, "sync single image timeout")
	gcrCmd.PersistentFlags().BoolVar(&gcrSyncOption.Kubeadm, "kubeadm", false, "sync kubeadm images(ignore namespace, use k8s.gcr.io)")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.BatchSize, "batch-size", 0, "batch size")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.BatchNumber, "batch-number", 0, "batch number")
	gcrCmd.PersistentFlags().BoolVar(&gcrSyncOption.OnlyDownloadManifests, "download-manifests", false, "only download manifests")
	gcrCmd.PersistentFlags().BoolVar(&gcrSyncOption.Report, "report", false, "report sync detail")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.ReportLevel, "report-level", 1, "report sync detail level")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.ReportName, "report-name", "kubeadm", "report name")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.TelegramApi, "telegram-api", "https://api.telegram.org", "telegram api address")
	gcrCmd.PersistentFlags().StringVar(&gcrSyncOption.TelegramToken, "telegram-token", "", "telegram bot token")
	gcrCmd.PersistentFlags().Int64Var(&gcrSyncOption.TelegramGroup, "telegram-group", 0, "telegram group id")
	gcrCmd.PersistentFlags().StringVar(&core.ManifestDir, "manifests", "manifests", "manifests storage dir")
}
