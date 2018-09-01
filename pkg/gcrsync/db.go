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
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/mritd/gcrsync/pkg/utils"

	bolt "go.etcd.io/bbolt"
)

var once sync.Once
var db *bolt.DB

func dbinit() *bolt.DB {
	once.Do(func() {
		var err error
		db, err = bolt.Open("gcr.db", 0600, nil)
		utils.CheckAndExit(err)
		db.Update(func(tx *bolt.Tx) error {
			_, err = tx.CreateBucketIfNotExists([]byte("gcr"))
			return err
		})
	})
	return db
}

func checkImage(imageName string) bool {

	imageExist := false

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gcr"))
		result := b.Get([]byte(imageName))
		if len(result) > 0 {
			imageExist = true
		}
		return nil
	})

	if !utils.CheckErr(err) {
		logrus.Errorf("Found imageName [%s] failed", imageName)
	}

	return imageExist
}

func putImage(imageName string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gcr"))
		return b.Put([]byte(imageName), []byte("true"))
	})
	if !utils.CheckErr(err) {
		logrus.Errorf("Failed to put imageName [%s]", imageName)
	}
}
