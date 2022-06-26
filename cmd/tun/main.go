package main

import (
	"fmt"
	"os"

	"github.com/byebyebruce/tun/tun"
	"github.com/spf13/cobra"
)

func main() {
	// server
	serverCmd := &cobra.Command{
		Use: "s",
	}
	s := serverCmd.Flags().String("listen", ":9900", "listen address")
	serverCmd.RunE = func(cmd *cobra.Command, args []string) error {
		server := tun.New(*s)
		return server.Run()
	}

	// client
	clientCmd := &cobra.Command{
		Use: "c localAddress",
	}
	c := clientCmd.Flags().String("server", "127.0.0.1:9900", "connect server")
	r := clientCmd.Flags().String("remote", ":9901", "remote address")
	clientCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("local address empty")
		}
		local := args[0]
		c, err := tun.NewClient(*c)
		if err != nil {
			return err
		}
		return c.Run(local, *r)
	}

	root := &cobra.Command{}
	root.AddCommand(serverCmd, clientCmd)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
