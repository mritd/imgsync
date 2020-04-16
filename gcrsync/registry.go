package gcrsync

import (
	"fmt"
	"io/ioutil"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/mritd/gcrsync/utils"
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
