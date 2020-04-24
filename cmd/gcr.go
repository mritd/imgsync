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
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.QueryLimit, "querylimit", core.DefaultLimit, "http query limit")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.Limit, "processlimit", core.DefaultLimit, "sync image limit")
	gcrCmd.PersistentFlags().DurationVar(&gcrSyncOption.Timeout, "timeout", core.DefaultSyncTimeout, "sync single image timeout")
	gcrCmd.PersistentFlags().BoolVar(&gcrSyncOption.Kubeadm, "kubeadm", false, "sync kubeadm images(ignore namespace, use k8s.gcr.io)")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.BatchSize, "batchsize", 0, "batch size")
	gcrCmd.PersistentFlags().IntVar(&gcrSyncOption.BatchNumber, "batchnumber", 0, "batch number")
}
