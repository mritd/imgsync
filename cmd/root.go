// Copyright © 2018 mritd <mritd1234@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/gcrsync"

	"github.com/spf13/cobra"
)

var debug bool
var prefix, proxy, dockerUser, dockerPassword string
var imageLimit int

var rootCmd = &cobra.Command{
	Use:   "gcrsync",
	Short: "A docker image sync tool for Google container registry (gcr.io)",
	Long: `
A docker image sync tool for Google container registry (gcr.io).`,
	Run: func(cmd *cobra.Command, args []string) {

		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		gcr := &gcrsync.Gcr{
			Prefix:         prefix,
			Proxy:          proxy,
			DockerUser:     dockerUser,
			DockerPassword: dockerPassword,
			ImageLimit:     imageLimit,
		}
		gcr.Init()
		gcr.Sync()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.PersistentFlags().StringVar(&prefix, "prefix", "gcrxio", "image prefix")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "gcr proxy")
	rootCmd.PersistentFlags().StringVar(&dockerUser, "user", "", "docker registry user")
	rootCmd.PersistentFlags().StringVar(&dockerPassword, "password", "", "docker registry user password")
	rootCmd.PersistentFlags().IntVar(&imageLimit, "limit", 100, "image sync limit(default 100)")
}
