/*
 * Copyright © 2018 mritd <mritd1234@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/json-iterator/go"

	"github.com/sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) dockerHubImages() []string {
	var images []string
	var val []struct {
		Name string
	}
	addr := fmt.Sprintf(DockerHubImage, g.DockerUser)
	for {
		req, _ := http.NewRequest("GET", addr, nil)
		resp, err := g.httpClient.Do(req)
		utils.CheckAndExit(err)
		if resp.StatusCode != http.StatusOK {
			utils.ErrorExit("Get docker hub images failed!", 1)
		}

		b, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		utils.CheckAndExit(err)

		_ = jsoniter.UnmarshalFromString(jsoniter.Get(b, "results").ToString(), &val)

		for _, v := range val {
			images = append(images, v.Name)
		}

		addr = jsoniter.Get(b, "next").ToString()
		if addr == "" {
			break
		}

	}
	return images
}

func (g *Gcr) dockerHubImageList() []string {

	var images []string
	dockerHubImages := g.dockerHubImages()

	logrus.Debugf("Number of docker hub images: %d", len(dockerHubImages))

	imgNameCh := make(chan string, 20)
	imgGetWg := new(sync.WaitGroup)
	imgGetWg.Add(len(dockerHubImages))

	for _, imageName := range dockerHubImages {

		tmpImageName := imageName

		go func() {
			defer func() {
				g.QueryLimit <- 1
				imgGetWg.Done()
			}()

			addr := fmt.Sprintf(DockerHubTags, g.DockerUser, tmpImageName)

			select {
			case <-g.QueryLimit:
				for {
					req, err := http.NewRequest("GET", addr, nil)
					utils.CheckAndExit(err)

					resp, err := g.httpClient.Do(req)
					utils.CheckAndExit(err)

					b, err := ioutil.ReadAll(resp.Body)
					utils.CheckAndExit(err)
					_ = resp.Body.Close()

					var val []struct {
						Name string
					}
					_ = jsoniter.UnmarshalFromString(jsoniter.Get(b, "results").ToString(), &val)

					for _, tag := range val {
						imgNameCh <- tmpImageName + ":" + tag.Name
					}

					addr = jsoniter.Get(b, "next").ToString()
					if addr == "" {
						break
					}
				}
			}

		}()
	}

	var imgReceiveWg sync.WaitGroup
	imgReceiveWg.Add(1)
	go func() {
		defer imgReceiveWg.Done()
		for {
			select {
			case imageName, ok := <-imgNameCh:
				if ok {
					images = append(images, imageName)
				} else {
					goto imgSetExit
				}
			}
		}
	imgSetExit:
	}()

	imgGetWg.Wait()
	close(imgNameCh)
	imgReceiveWg.Wait()
	return images

}
