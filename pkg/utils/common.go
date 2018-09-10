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

package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
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
		os.Exit(1)
	}
}

func ErrorExit(msg string, code int) {
	logrus.Errorln(msg)
	if code == 0 {
		code = 1
	}
	os.Exit(code)
}

func MustExec(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
}

func MustExecRtOut(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	return string(b)
}

func MustExecNoOut(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
}

func TryExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}
