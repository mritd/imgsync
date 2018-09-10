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

package gcrsync

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"

	"github.com/mritd/gcrsync/pkg/utils"
)

const (
	ChangeLog      = "CHANGELOG.md"
	GcrRegistryTpl = "gcr.io/%s/%s"
	GcrImages      = "https://gcr.io/v2/%s/tags/list"
	GcrImageTags   = "https://gcr.io/v2/%s/%s/tags/list"
	HubLogin       = "https://hub.docker.com/v2/users/login"
	HubRepos       = "https://hub.docker.com/v2/repositories/%s/?page_size=100"
	HubTags        = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=100"
)

func (g *Gcr) Sync() {

	gcrImages := g.gcrImageList()
	regImages := g.regImageList()

	logrus.Infof("Google container registry images total: %d", len(gcrImages))
	logrus.Infof("Docker registry images total: %d", len(regImages))

	for _, imageName := range regImages {
		if gcrImages[imageName] {
			logrus.Debugf("Image [%s] found, skip!", imageName)
			delete(gcrImages, imageName)
		}
	}

	logrus.Infof("Number of images waiting to be processed: %d", len(gcrImages))

	keys := make([]string, 0, len(gcrImages))
	for key := range gcrImages {
		keys = append(keys, key)
	}

	processWg := new(sync.WaitGroup)
	processWg.Add(len(keys))

	for _, imageName := range keys {
		tmpImageName := imageName
		go func() {
			defer func() {
				g.ProcessLimit <- 1
				processWg.Done()
			}()
			select {
			case <-g.ProcessLimit:
				g.Process(tmpImageName)
			}
		}()
	}

	// doc gen
	chgWg := new(sync.WaitGroup)
	chgWg.Add(1)
	go func() {
		defer chgWg.Done()

		var init bool
		updateInfo := ""
		for {
			select {
			case imageName, ok := <-g.update:
				if ok {
					if !init {
						updateInfo += fmt.Sprintf("### %s Update:\n\n", time.Now().Format("2006-01-02 15:04:05"))
						init = true
					}
					updateInfo += "- " + imageName + "\n"
				} else {
					goto ChangeLogDone
				}
			}
		}
	ChangeLogDone:
		if strings.TrimSpace(updateInfo) != "" {
			g.Commit(updateInfo)
		}
	}()

	processWg.Wait()
	close(g.update)
	chgWg.Wait()

}

func (g *Gcr) Monitor(count int) {

	if count == -1 {
		for {
			g.monitor()
		}
	} else {
		for i := 0; i < count; i++ {
			g.monitor()
		}
	}

}

func (g *Gcr) monitor() {

	select {
	case <-time.Tick(5 * time.Second):
		gcrImages := g.gcrImageList()
		regImages := g.regImageList()

		gcrSum := len(gcrImages)
		regSum := len(regImages)

		logrus.Debugf("Google container registry images total: %d", gcrSum)
		logrus.Debugf("Docker registry images total: %d", regSum)

		for _, imageName := range regImages {
			if gcrImages[imageName] {
				delete(gcrImages, imageName)
			}
		}

		processSum := len(gcrImages)

		logrus.Infof("Gcr images: %d    Registry images: %d    Waiting process: %d", gcrSum, regSum, processSum)
		logrus.Infoln(gcrImages)
	}

}

func (g *Gcr) Init() {

	logrus.Infoln("Init http client.")
	g.httpClient = &http.Client{
		Timeout: g.HttpTimeOut,
	}
	if g.Proxy != "" {
		p := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(g.Proxy)
		}
		g.httpClient.Transport = &http.Transport{Proxy: p}
	}

	logrus.Infoln("Init docker client.")
	dockerClient, err := client.NewEnvClient()
	utils.CheckAndExit(err)
	g.dockerClient = dockerClient

	logrus.Infoln("Init limit channel.")
	for i := 0; i < cap(g.QueryLimit); i++ {
		g.QueryLimit <- 1
	}
	for i := 0; i < cap(g.ProcessLimit); i++ {
		g.ProcessLimit <- 1
	}

	logrus.Infoln("Init update channel.")
	g.update = make(chan string, 20)

	logrus.Infoln("Init commit repo.")
	if g.GithubToken == "" {
		utils.ErrorExit("Github Token is blank!", 1)
	}
	g.commitURL = "https://" + g.GithubToken + "@github.com/" + g.GithubRepo + ".git"
	g.Clone()

	logrus.Infoln("Init success...")
}
