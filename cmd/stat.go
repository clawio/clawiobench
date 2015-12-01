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
	"text/tabwriter"
)

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

	in := &pb.StatReq{}
	in.AccessToken = token
	in.Path = args[0]
	in.Children = childrenFlag

	ctx := context.Background()

	res, err := c.Stat(ctx, in)
	if err != nil {
		return fmt.Errorf("Cannot stat resource: " + err.Error())
	}

	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer tabWriter.Flush()

	fmt.Fprintln(tabWriter, "ID\tPath\tContainer\tSize\tModified\tPermissions\tETag\tMime\tChecksum")

	fmt.Fprintf(tabWriter, "%s\t%s\t%t\t%d\t%d\t%d\t%s\t%s\t%s\n",
		res.Id, res.Path, res.IsContainer, res.Size, res.Modified, res.Permissions, res.Etag, res.MimeType, res.Checksum)

	for _, child := range res.GetChildren() {
		fmt.Fprintf(tabWriter, "%s\t%s\t%t\t%d\t%d\t%d\t%s\t%s\t%s\n",
			child.Id, child.Path, child.IsContainer, child.Size, child.Modified, child.Permissions, child.Etag, child.MimeType, child.Checksum)
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

}
