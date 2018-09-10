package gcrsync

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) needProcessImages(images []string) []string {
	var needSyncImages []string
	var imgGetWg sync.WaitGroup
	imgGetWg.Add(len(images))
	imgNameCh := make(chan string, 20)

	for _, imageName := range images {
		tmpImageName := imageName
		go func() {
			defer func() {
				g.QueryLimit <- 1
				imgGetWg.Done()
			}()

			select {
			case <-g.QueryLimit:
				if !g.queryRegistryImage(tmpImageName) {
					imgNameCh <- tmpImageName
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case imageName, ok := <-imgNameCh:
				if ok {
					needSyncImages = append(needSyncImages, imageName)
				} else {
					goto imgSetExit
				}
			}
		}
	imgSetExit:
	}()

	imgGetWg.Wait()
	close(imgNameCh)
	return needSyncImages

}

func (g *Gcr) queryRegistryImage(imageName string) bool {
	imageInfo := strings.Split(imageName, ":")
	addr := fmt.Sprintf(RegistryTag, g.DockerUser, imageInfo[0], imageInfo[1])
	req, _ := http.NewRequest("GET", addr, nil)
	resp, err := g.httpClient.Do(req)
	if !utils.CheckErr(err) {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		logrus.Debugf("Image [%s] found, skip!", imageName)
		return true
	} else {
		return false
	}
}
