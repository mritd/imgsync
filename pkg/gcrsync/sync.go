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
	"os"
	"sync"
	"time"

	"github.com/mritd/gcrsync/pkg/utils"
)

const (
	GcrImages    = "https://gcr.io/v2/google_containers/tags/list"
	GcrImageTags = "https://gcr.io/v2/google_containers/%s/tags/list"
	HubLogin     = "https://hub.docker.com/v2/users/login/"
	HubRepos     = "https://hub.docker.com/v2/repositories/%s/?page_size=10000"
	HubTags      = "https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=10000 "
)

func (g *Gcr) Sync() {

	images := g.gcrImageList()
	chgwg := new(sync.WaitGroup)
	chgwg.Add(1)

	// doc gen
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
					updateInfo += fmt.Sprintf("### %s Update:\n\n", time.Now().Format("2006-01-02 15:04:05"))
					init = true
				}
				updateInfo += "- " + imageName + "\n"
			}
		}
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
	close(g.update)
	chgwg.Wait()

}
