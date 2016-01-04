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
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"
)

var countFlag int
var bsFlag int
var checksumFlag string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Benchmarks the uploading process using different object sizes",
	RunE:  upload,
	Long: `This benchmark test will measure the upload performance.

The object size is the result of block size x count. This is the same
approach used by dd.`,
}

// createFile is a substitute for dd
// char is the character to insert
// count is the number of blocks
// bs is the block size: how many bytes are we going to write flush every round.
func createFile(fn, char string, count, bs int) (*os.File, error) {
	var fd *os.File
	if fn == "" {
		tf, err := ioutil.TempFile("", "CLAWIOBENCH-")
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
	if len(args) != 1 {
		cmd.Help()
		return nil
	}

	if concurrencyFlag > probesFlag {
		concurrencyFlag = probesFlag
	}
	if concurrencyFlag == 0 {
		concurrencyFlag++
	}

	token, err := getToken()
	if err != nil {
		log.Error(err)
		return err
	}

	fns := []string{}
	for i := 0; i < probesFlag; i++ {
		fd, err := createFile(fmt.Sprintf("testfile-%d", i), "1", countFlag, bsFlag)
		if err != nil {
			log.Error(err)
			return err
		}
		defer func() {
			fd.Close()
			os.RemoveAll(fd.Name())
		}()
		fns = append(fns, fd.Name())
	}

	benchStart := time.Now()

	total := 0
	errorProbes := 0

	errChan := make(chan error)
	resChan := make(chan string)
	doneChan := make(chan bool)
	limitChan := make(chan int, concurrencyFlag)

	for i := 0; i < concurrencyFlag; i++ {
		limitChan <- 1
	}

	var bar *pb.ProgressBar
	if progressBar {
		fmt.Printf("File size is %d megabytes\n", countFlag*bsFlag/1024/1024)
		bar = pb.StartNew(probesFlag)
	}

	for i := 0; i < probesFlag; i++ {
		go func(fn string) {
			<-limitChan
			defer func() {
				limitChan <- 1
			}()

			// open again the file
			lfd, err := os.Open(fn)
			if err != nil {
				errChan <- err
			}
			c := &http.Client{} // connections are reused if we reuse the client
			// PUT will close the fd
			// is it possible that the HTTP client is reusing connections so is being blocked?
			req, err := http.NewRequest("PUT", dataAddr+args[0], lfd)
			if err != nil {
				errChan <- err
				return
			}

			req.Header.Add("Content-Type", "application/octet-stream")
			req.Header.Add("Authorization", "Bearer "+token)
			req.Header.Add("CIO-Checksum", checksumFlag)

			res, err := c.Do(req)
			if err != nil {
				errChan <- err
				return
			}

			err = res.Body.Close()
			if err != nil {
				errChan <- err
				return
			}

			if res.StatusCode != 201 {
				err := fmt.Errorf("Request failed with status code %d", res.StatusCode)
				errChan <- err
				return
			}

			doneChan <- true
			resChan <- ""
			return
		}(fns[i])
	}

	for {
		select {
		case _ = <-doneChan:
			total++
			if progressBar {
				bar.Increment()
			}
		case _ = <-resChan:
		case err := <-errChan:
			log.Error(err)
			errorProbes++
			total++
			if progressBar {
				bar.Increment()
			}
		}

		if total == probesFlag {
			break
		}
	}

	if progressBar {
		bar.Finish()
	}

	numberRequests := probesFlag
	concurrency := concurrencyFlag
	totalTime := time.Since(benchStart).Seconds()
	failedRequests := errorProbes
	frequency := float64(numberRequests-failedRequests) / totalTime
	period := float64(1 / frequency)
	volume := numberRequests * countFlag * bsFlag / 1024 / 1024
	throughput := float64(volume) / totalTime
	data := [][]string{
		{"#NUMBER", "CONCURRENCY", "TIME", "FAILED", "FREQ", "PERIOD", "VOLUME", "THROUGHPUT"},
		{fmt.Sprintf("%d", numberRequests), fmt.Sprintf("%d", concurrency), fmt.Sprintf("%f", totalTime), fmt.Sprintf("%d", failedRequests), fmt.Sprintf("%f", frequency), fmt.Sprintf("%f", period), fmt.Sprintf("%d", volume), fmt.Sprintf("%f", throughput)},
	}
	w := csv.NewWriter(output)
	w.Comma = ' '
	for _, d := range data {
		if err := w.Write(d); err != nil {
			return err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
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
	uploadCmd.Flags().StringVar(&checksumFlag, "checksum", "", "The checksum for the file")

}
