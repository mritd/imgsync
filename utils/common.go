package utils

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func CheckErr(err error) bool {
	if err != nil {
		logrus.Errorln(err)
		return false
	}

	return true
}

func CheckAndExit(err error) {
	if !CheckErr(err) {
		panic(err)
	}
}

func ErrorExit(msg string, code int) {
	logrus.Errorln(msg)
	if code == 0 {
		code = 1
	}
	os.Exit(code)
}

func GitCmd(dir string, arg ...string) {
	cmd := exec.Command("git", arg...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
}

func SliceDiff(slice1, slice2 []string) (diffslice []string) {
	for _, v := range slice1 {
		if !inSliceIface(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}

func inSliceIface(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}
