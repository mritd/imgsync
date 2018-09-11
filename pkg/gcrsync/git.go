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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/json-iterator/go"

	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) Commit(images []string) {

	repoDir := strings.Split(g.GithubRepo, "/")[1]
	repoChangeLog := filepath.Join(repoDir, ChangeLog)
	repoUpdateFile := filepath.Join(repoDir, g.NameSpace)

	var content []byte
	chgLog, err := os.Open(repoChangeLog)
	if utils.CheckErr(err) {
		defer chgLog.Close()
		content, err = ioutil.ReadAll(chgLog)
		utils.CheckAndExit(err)
	}

	chgLog, err = os.OpenFile(repoChangeLog, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer chgLog.Close()

	updateInfo := fmt.Sprintf("### %s Update:\n\n", time.Now().Format("2006-01-02 15:04:05"))
	for _, imageName := range images {
		updateInfo += "- " + fmt.Sprintf(GcrRegistryTpl, g.NameSpace, imageName) + "\n"
	}
	chgLog.WriteString(updateInfo + string(content))

	var synchronizedImages []string
	updateFile, err := os.Open(repoUpdateFile)
	if utils.CheckErr(err) {
		defer updateFile.Close()
		content, err = ioutil.ReadAll(updateFile)
		utils.CheckAndExit(err)
	}
	utils.CheckAndExit(jsoniter.Unmarshal(content, &synchronizedImages))
	synchronizedImages = append(synchronizedImages, images...)
	sort.Strings(synchronizedImages)
	buf, err := jsoniter.MarshalIndent(synchronizedImages, "", "    ")
	utils.CheckAndExit(err)
	newUpdateFile, err := os.OpenFile(repoUpdateFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer newUpdateFile.Close()
	newUpdateFile.Write(buf)

	utils.GitCmd(repoDir, "config", "--global", "push.default", "simple")
	utils.GitCmd(repoDir, "config", "--global", "user.email", "gcrsync@mritd.me")
	utils.GitCmd(repoDir, "config", "--global", "user.name", "gcrsync")
	utils.GitCmd(repoDir, "add", ".")
	utils.GitCmd(repoDir, "commit", "-m", g.CommitMsg)
	utils.GitCmd(repoDir, "push", "--force", g.commitURL, "master")

}

func (g *Gcr) Clone() {
	os.RemoveAll(strings.Split(g.GithubRepo, "/")[1])
	utils.GitCmd("", "clone", g.commitURL)
}
