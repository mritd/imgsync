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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"

	"github.com/mritd/gcrsync/pkg/utils"
)

const (
	GcrRegistryTpl = "gcr.io/%s/%s"
	GcrImages      = "https://gcr.io/v2/%s/tags/list"
	GcrImageTags   = "https://gcr.io/v2/%s/%s/tags/list"
	HubLogin       = "https://hub.docker.com/v2/users/login/"
	HubRepos       = "https://hub.docker.com/v2/repositories/%s/?page_size=10000"
	HubTags        = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=10000"
)

func (g *Gcr) Sync() {

	gcrImages := g.gcrImageList()
	regImages := g.regImageList()

	logrus.Infof("Google container registry images total: %d", len(gcrImages))

	for _, imageName := range regImages {
		if gcrImages[imageName] {
			logrus.Debugf("Image [%s] found, skip!", imageName)
			delete(gcrImages, imageName)
		}
	}

	logrus.Infof("Number of images waiting to be processed: %d", len(gcrImages))

	var batchNum int
	if len(gcrImages) < g.ProcessLimit {
		g.ProcessLimit = len(gcrImages)
		batchNum = 1
	} else {
		batchNum = len(gcrImages) / g.ProcessLimit
	}

	logrus.Infof("Image process batchNum: %d", batchNum)

	keys := make([]string, 0, len(gcrImages))
	for key := range gcrImages {
		keys = append(keys, key)
	}

	processWg := new(sync.WaitGroup)
	processWg.Add(g.ProcessLimit)

	for i := 0; i < g.ProcessLimit; i++ {

		var tmpImages []string

		if i+1 == g.ProcessLimit {
			tmpImages = keys[i*batchNum:]
		} else {
			tmpImages = keys[i*batchNum : (i+1)*batchNum]
		}

		go func() {
			defer processWg.Done()
			for _, imageName := range tmpImages {
				g.Process(imageName)
			}
		}()
	}

	// doc gen
	chgWg := new(sync.WaitGroup)
	chgWg.Add(1)
	go func() {
		defer chgWg.Done()

		var contents []byte
		chglog, err := os.Open("CHANGELOG.md")
		if utils.CheckErr(err) {
			contents, _ = ioutil.ReadAll(chglog)
			chglog.Close()
		}

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
		chglog, err = os.OpenFile("CHANGELOG.md", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		utils.CheckAndExit(err)
		defer chglog.Close()
		newContents := updateInfo + "\n" + string(contents)
		chglog.WriteString(newContents)
	}()

	processWg.Wait()
	close(g.update)
	chgWg.Wait()

}

func (g *Gcr) Init() {

	logrus.Infoln("Init http client.")
	var httpClient *http.Client
	if g.Proxy != "" {
		p := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(g.Proxy)
		}
		transport := &http.Transport{Proxy: p}
		httpClient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		}
	} else {
		httpClient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	g.httpClient = httpClient

	logrus.Infoln("Init docker client.")
	dockerClient, err := client.NewEnvClient()
	utils.CheckAndExit(err)
	g.dockerClient = dockerClient

	logrus.Infoln("Init update channel.")
	g.update = make(chan string, 20)

	logrus.Infoln("Init success...")
}
