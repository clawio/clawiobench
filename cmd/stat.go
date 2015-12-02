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
	pb "github.com/clawio/clawiobench/proto/metadata"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

var childrenFlag bool

var statCmd = &cobra.Command{
	Use:   "stat <path>",
	Short: "Stat a resource",
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

	wg := sync.WaitGroup{}

	for i := 0; i < probesFlag; i++ {
		wg.Add(1)
		if concurrentFlag {
			go doStat(c, token, args[0], &wg)
		} else {
			doStat(c, token, args[0], &wg)
		}
	}

	wg.Wait()

	benchEnd := time.Since(benchStart)
	fmt.Printf("Total time: %f s\n", benchEnd.Seconds())
	fmt.Printf("Average stat rate: %f ops/s\n", float64(probesFlag)/benchEnd.Seconds())

	return nil
}

func doStat(c pb.MetaClient, path, token string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info("START")
	in := &pb.StatReq{}
	in.AccessToken = token
	in.Path = path
	in.Children = childrenFlag

	ctx := context.Background()

	_, err := c.Stat(ctx, in)
	if err != nil {
		log.Errorf("Cannot stat resource: " + err.Error())
	}
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
