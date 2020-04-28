package cmd

import (
	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var kNativeSyncOption core.SyncOption

var kNativeCmd = &cobra.Command{
	Use:   "knative",
	Short: "Sync kNative images",
	Long: `
Sync kNative images.`,
	PreRun: prerun,
	Run: func(cmd *cobra.Command, args []string) {
		boot("kNative", &kNativeSyncOption)
	},
}

func init() {
	rootCmd.AddCommand(kNativeCmd)
	kNativeCmd.PersistentFlags().StringVar(&kNativeSyncOption.User, "user", "", "docker hub user")
	kNativeCmd.PersistentFlags().StringVar(&kNativeSyncOption.Password, "password", "", "docker hub user password")
	kNativeCmd.PersistentFlags().IntVar(&kNativeSyncOption.QueryLimit, "query-limit", core.DefaultLimit, "http query limit")
	kNativeCmd.PersistentFlags().IntVar(&kNativeSyncOption.Limit, "process-limit", core.DefaultLimit, "sync image limit")
	kNativeCmd.PersistentFlags().DurationVar(&kNativeSyncOption.Timeout, "timeout", core.DefaultSyncTimeout, "sync single image timeout")
	kNativeCmd.PersistentFlags().IntVar(&kNativeSyncOption.BatchSize, "batch-size", 0, "batch size")
	kNativeCmd.PersistentFlags().IntVar(&kNativeSyncOption.BatchNumber, "batch-number", 0, "batch number")
	kNativeCmd.PersistentFlags().BoolVar(&kNativeSyncOption.OnlyDownloadManifests, "download-manifests", false, "only download manifests")
	kNativeCmd.PersistentFlags().BoolVar(&kNativeSyncOption.Report, "report", false, "report sync detail")
	kNativeCmd.PersistentFlags().IntVar(&kNativeSyncOption.ReportLevel, "report-level", 1, "report sync detail level")
	kNativeCmd.PersistentFlags().StringVar(&kNativeSyncOption.ReportFile, "report-file", "imgsync_report", "report sync detail file")
	kNativeCmd.PersistentFlags().StringVar(&core.ManifestDir, "manifests", "manifests", "manifests storage dir")
}
