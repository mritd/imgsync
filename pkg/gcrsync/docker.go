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
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/json-iterator/go"

	"github.com/sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/utils"

	"github.com/docker/docker/api/types"
)

func (g *Gcr) Process(imageName string) {

	logrus.Infof("Process image: %s", imageName)

	ctx := context.Background()
	oldImageName := fmt.Sprintf(GcrRegistryTpl, g.NameSpace, imageName)
	newImageName := "docker.io/" + g.DockerUser + "/" + imageName

	if !g.TestMode {

		// pull image
		r, err := g.dockerClient.ImagePull(ctx, oldImageName, types.ImagePullOptions{})
		if !utils.CheckErr(err) {
			logrus.Errorf("Failed to pull image: %s", oldImageName)
			return
		}
		_, _ = io.Copy(ioutil.Discard, r)
		_ = r.Close()
		logrus.Infof("Pull image: %s success.", oldImageName)

		// tag it
		err = g.dockerClient.ImageTag(ctx, oldImageName, newImageName)
		if !utils.CheckErr(err) {
			logrus.Errorf("Failed to tag image [%s] ==> [%s]", oldImageName, newImageName)
			return
		}
		logrus.Infof("Tag image: %s success.", oldImageName)

		// push image
		authConfig := types.AuthConfig{
			Username: g.DockerUser,
			Password: g.DockerPassword,
		}
		encodedJSON, err := jsoniter.Marshal(authConfig)
		if !utils.CheckErr(err) {
			logrus.Errorln("Failed to marshal docker config")
			return
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		r, err = g.dockerClient.ImagePush(ctx, newImageName, types.ImagePushOptions{RegistryAuth: authStr})
		if !utils.CheckErr(err) {
			logrus.Errorf("Failed to push image: %s", newImageName)
			return
		}
		_, _ = io.Copy(ioutil.Discard, r)
		_ = r.Close()
		logrus.Infof("Push image: %s success.", newImageName)

		// clean image
		_, _ = g.dockerClient.ImageRemove(ctx, oldImageName, types.ImageRemoveOptions{})
		_, _ = g.dockerClient.ImageRemove(ctx, newImageName, types.ImageRemoveOptions{})
		logrus.Debugf("Remove image: %s success.", oldImageName)

	}
	g.update <- imageName
	logrus.Debugln("Done.")

}
