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
	"os"
)

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Create user home directory",
	Run:   home,
}

func home(cmd *cobra.Command, args []string) {

	token, err := getToken()
	if err != nil {
		fmt.Println("Authentication required")
		os.Exit(1)
	}

	addr := os.Getenv("CLAWIO_BENCH_META_ADDR")

	con, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("Cannot connect to server " + addr)
		os.Exit(1)
	}

	defer con.Close()

	c := pb.NewMetaClient(con)

	in := &pb.HomeReq{}
	in.AccessToken = token

	ctx := context.Background()

	_, err = c.Home(ctx, in)
	if err != nil {
		fmt.Println("Cannot create homedir: " + err.Error())
		os.Exit(1)
	}

	fmt.Println("Home directory created")
}

func init() {
	RootCmd.AddCommand(homeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// homeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// homeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
