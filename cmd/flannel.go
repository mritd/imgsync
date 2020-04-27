package cmd

import (
	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var flSyncOption core.SyncOption

var flannelCmd = &cobra.Command{
	Use:   "flannel",
	Short: "Sync flannel images",
	Long: `
Sync flannel images.`,
	PreRun: prerun,
	Run: func(cmd *cobra.Command, args []string) {
		boot("flannel", &flSyncOption)
	},
}

func init() {
	rootCmd.AddCommand(flannelCmd)
	flannelCmd.PersistentFlags().StringVar(&flSyncOption.User, "user", "", "docker hub user")
	flannelCmd.PersistentFlags().StringVar(&flSyncOption.Password, "password", "", "docker hub user password")
	flannelCmd.PersistentFlags().DurationVar(&flSyncOption.Timeout, "timeout", core.DefaultSyncTimeout, "sync single image timeout")
	flannelCmd.PersistentFlags().IntVar(&flSyncOption.Limit, "process-limit", core.DefaultLimit, "sync image limit")
	flannelCmd.PersistentFlags().BoolVar(&flSyncOption.OnlyDownloadManifests, "download-manifests", false, "only download manifests")
	flannelCmd.PersistentFlags().BoolVar(&gcrSyncOption.Report, "report", false, "report sync detail")
	flannelCmd.PersistentFlags().IntVar(&gcrSyncOption.ReportLevel, "report-level", 1, "report sync detail level")
	flannelCmd.PersistentFlags().StringVar(&gcrSyncOption.ReportFile, "report-file", "imgsync_report", "report sync detail file")
	flannelCmd.PersistentFlags().StringVar(&core.ManifestDir, "manifests", "manifests", "manifests storage dir")
}
