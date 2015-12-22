package cmd

import (
	"fmt"
	pb "github.com/clawio/clawiobench/proto/auth"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <username> <password>",
	Short: "Login into ClawIO",
	Run:   login,
}

func login(cmd *cobra.Command, args []string) {

	if len(args) != 2 {
		cmd.Help()
		os.Exit(1)
	}

	con, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		log.Error(err)
		fmt.Println("Cannot connect to authentication unit")
		os.Exit(1)
	}

	defer con.Close()

	c := pb.NewAuthClient(con)

	in := &pb.AuthRequest{}
	in.Username = args[0]
	in.Password = args[1]

	ctx := context.Background()

	res, err := c.Authenticate(ctx, in)
	if err != nil {
		if grpc.Code(err) == codes.Unauthenticated {
			fmt.Println("Invalid username or password")
			os.Exit(1)
		}
		fmt.Println("Cannot connect to authentication unit")
		os.Exit(1)
	}

	// Save token into $HOME/.clawiobench/credentials
	u, err := user.Current()
	if err != nil {
		log.Error(err)
		fmt.Println("Cannot access your home directory")
		os.Exit(1)
	}

	err = os.MkdirAll(path.Join(u.HomeDir, ".clawiobench"), 0755)
	if err != nil {
		log.Error(err)
		fmt.Println("Cannot create $HOME/.clawiobench configuration directory")
		os.Exit(1)
	}

	err = ioutil.WriteFile(path.Join(u.HomeDir, ".clawiobench", "credentials"), []byte(res.Token), 0644)
	if err != nil {
		log.Error(err)
		fmt.Println("Cannot save credentials into $HOME/.clawiobench/credentials")
		os.Exit(1)
	}

	fmt.Println("You are logged in as " + in.Username)
	os.Exit(0)
}

func init() {
	RootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
