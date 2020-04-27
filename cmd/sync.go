package cmd

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mritd/imgsync/core"
	"github.com/spf13/cobra"
)

var syncOption core.SyncOption

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync single image",
	Long: `
Sync single image.`,
	PreRun: prerun,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
			return
		}
		var repo, user, name, tag string
		ss := strings.Split(args[0], ":")
		if len(ss) == 1 {
			tag = "latest"
		} else {
			tag = ss[len(ss)-1]
		}
		ss = strings.Split(ss[0], "/")
		switch len(ss) {
		case 1:
			name = ss[0]
		case 2:
			repo = ss[0]
			name = ss[1]
		case 3:
			repo = ss[0]
			user = ss[1]
			name = ss[2]
		default:
			logrus.Fatalf("image name format error: %s", args[0])
		}
		core.SyncImages(context.Background(), core.Images{core.Image{
			Repo: repo,
			User: user,
			Name: name,
			Tag:  tag,
		}}, &syncOption)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVar(&syncOption.User, "user", "", "docker hub user")
	syncCmd.PersistentFlags().StringVar(&syncOption.Password, "password", "", "docker hub user password")
	syncCmd.PersistentFlags().StringVar(&syncOption.NameSpace, "namespace", "google-containers", "google container registry namespace")
	syncCmd.PersistentFlags().DurationVar(&syncOption.Timeout, "timeout", core.DefaultSyncTimeout, "sync single image timeout")
	syncCmd.PersistentFlags().BoolVar(&syncOption.OnlyDownloadManifests, "download-manifests", false, "only download manifests")
	syncCmd.PersistentFlags().StringVar(&core.ManifestDir, "manifests", "manifests", "manifests storage dir")
}
