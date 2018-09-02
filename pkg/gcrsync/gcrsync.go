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

	"github.com/docker/docker/client"

	"github.com/Sirupsen/logrus"

	"github.com/json-iterator/go"

	"github.com/mritd/gcrsync/pkg/utils"
)

const (
	GcrImages    = "https://gcr.io/v2/google_containers/tags/list"
	GcrImageTags = "https://gcr.io/v2/google_containers/%s/tags/list"
)

type Image struct {
	Name string
	Tags []string
}

type Gcr struct {
	Proxy          string
	Prefix         string
	DockerUser     string
	DockerPassword string
	ImageLimit     int
	gcloudClient   *http.Client
	dockerClient   *client.Client
	update         chan string
}

func (g *Gcr) Sync() {
	defer db.Close()

	images := g.imageList()
	chgwg := new(sync.WaitGroup)
	chgwg.Add(1)
	go func() {
		defer chgwg.Done()
		var init bool
		chglog, err := os.OpenFile("CHANGELOG.md", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		utils.CheckAndExit(err)
		defer chglog.Close()
		contents, err := ioutil.ReadAll(chglog)
		utils.CheckAndExit(err)
		updateInfo := ""
		for {
			select {
			case imageName := <-g.update:
				if !init {
					updateInfo += fmt.Sprintf("\n### %s Update:\n\n", time.Now().Format("2006-01-02 15:04:05"))
					init = true
				}
				updateInfo += "- " + imageName + "\n"
			}
		}
		newContents := updateInfo + string(contents)
		chglog.WriteString(newContents)
	}()

	if len(images) <= 10 {
		g.Process(images)
	} else {
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
		close(g.update)
	}
	chgwg.Wait()

}

func (g *Gcr) Init() {
	logrus.Debugln("Init google container client.")
	var gclient *http.Client
	if g.Proxy != "" {
		p := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(g.Proxy)
		}
		transport := &http.Transport{Proxy: p}
		gclient = &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		}
	} else {
		gclient = &http.Client{
			Timeout: 5 * time.Second,
		}
	}
	g.gcloudClient = gclient

	logrus.Debugln("Init docker client.")
	dclient, err := client.NewEnvClient()
	utils.CheckAndExit(err)
	g.dockerClient = dclient

	logrus.Debugln("Init bbolt database.")
	dbinit()

	logrus.Debugln("Init update channel.")
	g.update = make(chan string, 10)

	logrus.Debugln("Init success...")
}

func (g *Gcr) imageList() []Image {

	var images []Image

	publicImages := g.publicImages()
	for _, imageName := range publicImages {
		req, err := http.NewRequest("GET", fmt.Sprintf(GcrImageTags, imageName), nil)
		utils.CheckAndExit(err)

		resp, err := g.gcloudClient.Do(req)
		utils.CheckAndExit(err)

		b, err := ioutil.ReadAll(resp.Body)
		utils.CheckAndExit(err)
		resp.Body.Close()

		var tags []string
		jsoniter.UnmarshalFromString(jsoniter.Get(b, "tags").ToString(), &tags)

		logrus.Debugf("Found image [%s] tags: %s", imageName, tags)

		images = append(images, Image{
			Name: GcrRegistryPrefix + imageName,
			Tags: tags,
		})

	}
	return images
}

func (g *Gcr) publicImages() []string {

	req, err := http.NewRequest("GET", GcrImages, nil)
	utils.CheckAndExit(err)

	resp, err := g.gcloudClient.Do(req)
	utils.CheckAndExit(err)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	utils.CheckAndExit(err)

	var images []string
	jsoniter.UnmarshalFromString(jsoniter.Get(b, "child").ToString(), &images)
	return images
}
