// Copyright © 2015 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

var countFlag int
var bsFlag int
var probesFlag int
var concurrentFlag bool

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Benchmarks the uploading process using different object sizes",
	RunE:  upload,
	Long: `This benchmark test will measure the upload performance.

The object size is the result of block size x count. This is the same
approach used by dd.

By default, uploads will be done sequentially. If you want your
uploads to be performed in parallel set the --concurrent flag.`,
}

// createFile is a substitute for dd
// char is the character to insert
// count is the number of blocks
// bs is the block size: how many bytes are we going to write flush every round.
func createFile(fn, char string, count, bs int) (*os.File, error) {
	var fd *os.File
	if fn == "" {
		tf, err := ioutil.TempFile("", "clawiobench-")
		if err != nil {
			return nil, err
		}
		fd = tf
	} else {
		tf, err := os.Create(path.Join(os.TempDir(), fn))
		if err != nil {
			log.Error(err)
			return nil, err
		}
		fd = tf
	}

	// if char is 1 byte then the buffer size will be equal to bs
	buffer := bytes.Repeat([]byte(char), bs)

	for i := 0; i < count; i++ {
		_, err := fd.Write(buffer)
		if err != nil {
			return nil, err
		}
	}

	return fd, nil
}

func upload(cmd *cobra.Command, args []string) error {

	token, err := getToken()
	if err != nil {
		log.Error(err)
		return err
	}

	fd, err := createFile("", "1", countFlag, bsFlag) // 1GB file
	if err != nil {
		log.Error(err)
		return err
	}

	fn := fd.Name()
	fmt.Println("Test file is " + fn)
	fmt.Printf("File size is %d megabytes", countFlag*bsFlag/1024/1024)

	benchStart := time.Now()

	var wg sync.WaitGroup
	c := &http.Client{}

	for i := 0; i < probesFlag; i++ {
		wg.Add(1)
		if concurrentFlag {
			go doUpload(c, fn, token, &wg)
		} else {
			doUpload(c, fn, token, &wg)
		}
	}

	wg.Wait()

	benchEnd := time.Since(benchStart)
	defer fd.Close()
	fmt.Printf("Total number of probes: %d\n", probesFlag)
	fmt.Printf("Data volume uploaded: %d megabytes\n", countFlag*bsFlag*probesFlag/1024/1024)
	fmt.Printf("Total upload time: %f s\n", benchEnd.Seconds())
	fmt.Printf("Average time per upload: %f s\n", benchEnd.Seconds()/float64(probesFlag))
	fmt.Printf("Average upload speed: %f megabytes/s\n", float64((countFlag*bsFlag*probesFlag/1024/1024))/benchEnd.Seconds())

	return nil
}

func doUpload(c *http.Client, fn, token string, wg *sync.WaitGroup) {
	defer wg.Done()

	workerID := uuid.New()
	time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
	log.WithField("workerid", workerID).Info("START")
	// open again the file
	lfd, err := os.Open(fn)
	if err != nil {
		log.Error(err)
	}
	// PUT will close the fd
	// is it possible that the HTTP client is reusing connections so is being blocked?
	req, err := http.NewRequest("PUT", dataAddr+"/local/users/d/demo/benchmark-file", lfd)
	if err != nil {
		log.Error(err)
	}

	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Authorization", "Bearer "+token)
	//req.Header.Add("CIO-Checksum", checksumFlag)

	log.WithField("workerid", workerID).Info("START REQ")
	res, err := c.Do(req)
	if err != nil {
		log.Error(err)
	}

	log.WithField("workerid", workerID).Info("END REQ")
	err = res.Body.Close()
	if err != nil {
		log.Error(err)
	}

	if res.StatusCode == 412 {
		err := fmt.Errorf("Object %s was corrupted during upload and server did not save it\n")
		log.Error(err)
	}
	log.WithField("workerid", workerID).Info("FINISH")
}

func init() {
	RootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	uploadCmd.Flags().IntVar(&countFlag, "count", 1024, "The number of blocks of the file")
	uploadCmd.Flags().IntVar(&bsFlag, "bs", 1024, "The number of bytes of each block")
	uploadCmd.Flags().IntVar(&probesFlag, "probes", 1, "The number of uploads")
	uploadCmd.Flags().BoolVar(&concurrentFlag, "concurrent", false, "If set uploads will be performed concurrently")

}
