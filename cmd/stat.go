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
	"encoding/csv"
	"fmt"
	br "github.com/cheggaaa/pb"
	pb "github.com/clawio/clawiobench/proto/metadata"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

var childrenFlag bool

var statCmd = &cobra.Command{
	Use:   "stat <path>",
	Short: "Benchmark getting resource information using stat",
	RunE:  stat,
}

func stat(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		cmd.Help()
		return nil
	}

	token, err := getToken()
	if err != nil {
		return err
	}

	con, err := grpc.Dial(metaAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer con.Close()

	c := pb.NewMetaClient(con)

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

	var bar *br.ProgressBar
	if progressBar {
		bar = br.StartNew(probesFlag)
	}

	for i := 0; i < probesFlag; i++ {
		go func() {
			<-limitChan
			defer func() {
				limitChan <- 1
			}()
			in := &pb.StatReq{}
			in.AccessToken = token
			in.Path = args[0]
			in.Children = childrenFlag
			ctx := context.Background()
			_, err := c.Stat(ctx, in)
			if err != nil {
				errChan <- err
				return
			}
			doneChan <- true
			resChan <- ""
		}()
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

	data := [][]string{
		{"#NUMBER", "CONCURRENCY", "TIME", "FAILED", "FREQ", "PERIOD"},
		{fmt.Sprintf("%d", numberRequests), fmt.Sprintf("%d", concurrency), fmt.Sprintf("%f", totalTime), fmt.Sprintf("%d", failedRequests), fmt.Sprintf("%f", frequency), fmt.Sprintf("%f", period)},
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
	RootCmd.AddCommand(statCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	statCmd.Flags().BoolVarP(&childrenFlag, "children", "", false, "Show children objects inside container")
}
