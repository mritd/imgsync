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
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/gcrsync"

	"github.com/spf13/cobra"
)

var debug, test, monitor bool
var proxy, dockerUser, dockerPassword, nameSpace string
var githubRepo, githubToken, commitMsg string
var queryLimit, processLimit, monitorCount int
var httpTimeout time.Duration

var rootCmd = &cobra.Command{
	Use:   "gcrsync",
	Short: "A docker image sync tool for Google container registry (gcr.io)",
	Long: `
A docker image sync tool for Google container registry (gcr.io).`,
	Run: func(cmd *cobra.Command, args []string) {

		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		gcr := &gcrsync.Gcr{
			Proxy:          proxy,
			DockerUser:     dockerUser,
			DockerPassword: dockerPassword,
			NameSpace:      nameSpace,
			TestMode:       test,
			QueryLimit:     make(chan int, queryLimit),
			ProcessLimit:   make(chan int, processLimit),
			HttpTimeOut:    httpTimeout,
		}
		gcr.Init()
		if !monitor {
			gcr.Sync()
		} else {
			gcr.Monitor(monitorCount)
		}
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
	rootCmd.PersistentFlags().BoolVar(&test, "test", false, "run test mode(only write changelog)")
	rootCmd.PersistentFlags().BoolVar(&monitor, "monitor", false, "monitor images sync detail")
	rootCmd.PersistentFlags().IntVar(&monitorCount, "monitorcount", -1, "monitor count")
	rootCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "http client proxy")
	rootCmd.PersistentFlags().StringVar(&dockerUser, "user", "", "docker registry user")
	rootCmd.PersistentFlags().StringVar(&dockerPassword, "password", "", "docker registry user password")
	rootCmd.PersistentFlags().StringVar(&nameSpace, "namespace", "google_containers", "google container registry namespace")
	rootCmd.PersistentFlags().IntVar(&queryLimit, "querylimit", 50, "http query limit")
	rootCmd.PersistentFlags().DurationVar(&httpTimeout, "httptimeout", 10*time.Second, "http request timeout")
	rootCmd.PersistentFlags().IntVar(&processLimit, "processlimit", 10, "image process limit")
	rootCmd.PersistentFlags().StringVar(&githubRepo, "githubrepo", "mritd/gcr", "github commit repo")
	rootCmd.PersistentFlags().StringVar(&githubToken, "githubtoken", "", "github commit token")
	rootCmd.PersistentFlags().StringVar(&commitMsg, "commitmsg", "Travis CI Auto Synchronized.", "github commit message")
}
