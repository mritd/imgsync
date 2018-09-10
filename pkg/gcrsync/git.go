package gcrsync

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mritd/gcrsync/pkg/utils"
)

func (g *Gcr) Commit(updateInfo string) {

	repoDir := strings.Split(g.GithubRepo, "/")[1]
	repoChangeLog := filepath.Join(repoDir, ChangeLog)

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
	chgLog.WriteString(updateInfo + string(content))
	utils.MustExec("git", "--git-dir="+repoDir+".git", "--force", g.commitURL)

}

func (g *Gcr) Clone() {
	utils.MustExec("git", "clone", g.commitURL)
}
