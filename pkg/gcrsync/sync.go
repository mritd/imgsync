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
	GcrRegistryPrefix = "gcr.io/%s/"
	GcrImages         = "https://gcr.io/v2/%s/tags/list"
	GcrImageTags      = "https://gcr.io/v2/%s/%s/tags/list"
	HubLogin          = "https://hub.docker.com/v2/users/login/"
	HubRepos          = "https://hub.docker.com/v2/repositories/%s/?page_size=10000"
	HubTags           = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=10000"
)

func (g *Gcr) Sync() {

	images := g.gcrImageList()

	// doc gen
	chgdone := make(chan int, 1)
	chgwg := new(sync.WaitGroup)
	chgwg.Add(1)
	go func() {
		defer chgwg.Done()

		var contents []byte
		chglog, err := os.Open("CHANGELOG.md")
		if !utils.CheckErr(err) {
			contents, _ = ioutil.ReadAll(chglog)
			chglog.Close()
		}

		var init bool
		updateInfo := ""
		for {
			select {
			case imageName := <-g.update:
				logrus.Infoln(imageName)
				if !init {
					updateInfo += fmt.Sprintf("### %s Update:\n\n", time.Now().Format("2006-01-02 15:04:05"))
					init = true
				}
				updateInfo += "- " + imageName + "\n"
			case <-chgdone:
				goto ChangeLogDone
			}
		}
	ChangeLogDone:
		chglog, err = os.OpenFile("CHANGELOG.md", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		newContents := updateInfo + "\n" + string(contents)
		chglog.WriteString(newContents)
	}()

	count := len(images) / 10
	if len(images)%10 > 0 {
		count++
	}

	var wg = new(sync.WaitGroup)
	wg.Add(count)

	for i := 0; i < count; i++ {
		x := i
		go func() {
			defer wg.Done()
			if x == count-1 {
				g.Process(images[x*10 : x*10+len(images)%10])
			} else {
				g.Process(images[x*10 : (x+1)*10])
			}
		}()
	}
	wg.Wait()
	chgdone <- 1
	chgwg.Wait()

}

func (g *Gcr) Init() {
	logrus.Debugln("Init http client.")
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

	logrus.Debugln("Init docker client.")
	dockerClient, err := client.NewEnvClient()
	utils.CheckAndExit(err)
	g.dockerClient = dockerClient

	logrus.Debugln("Init docker hub images.")
	g.hubImages()

	logrus.Debugln("Init update channel.")
	g.update = make(chan string, 20)

	logrus.Debugln("Init success...")
}
