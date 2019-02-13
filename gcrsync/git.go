package gcrsync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mritd/gcrsync/utils"
)

func (g *Gcr) Commit(images []string) {

	loc, _ := time.LoadLocation("Asia/Shanghai")

	repoDir := filepath.Join(strings.Split(g.GithubRepo, "/")[1], "changelog")
	if _, err := os.Stat(repoDir); err != nil {
		_ = os.MkdirAll(repoDir, 0755)
	}

	repoChangeLog := filepath.Join(repoDir, fmt.Sprintf(ChangeLog, time.Now().In(loc).Format("2006-01-02")))

	var content []byte
	chgLog, err := os.OpenFile(repoChangeLog, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	utils.CheckAndExit(err)
	defer func() {
		_ = chgLog.Close()
	}()

	updateInfo := fmt.Sprintf("### %s Update:\n\n", time.Now().In(loc).Format("2006-01-02 15:04:05"))
	for _, imageName := range images {
		updateInfo += "- " + fmt.Sprintf(GcrRegistryTpl, g.NameSpace, imageName) + "\n"
	}
	_, _ = chgLog.WriteString(updateInfo + string(content))

	utils.GitCmd(repoDir, "config", "--global", "push.default", "simple")
	utils.GitCmd(repoDir, "config", "--global", "user.email", "gcrsync@mritd.me")
	utils.GitCmd(repoDir, "config", "--global", "user.name", "gcrsync")
	utils.GitCmd(repoDir, "add", ".")
	utils.GitCmd(repoDir, "commit", "-m", g.CommitMsg)
	if !g.TestMode {
		utils.GitCmd(repoDir, "push", "--force", g.commitURL, "master")
	}

}

func (g *Gcr) Clone() {
	_ = os.RemoveAll(strings.Split(g.GithubRepo, "/")[1])
	utils.GitCmd("", "clone", g.commitURL)
}
