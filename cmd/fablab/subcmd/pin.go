package subcmd

import (
	"fmt"

	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(pinCmd)
	RootCmd.AddCommand(unpinCmd)
}

var pinCmd = &cobra.Command{
	Use:   "pin <instance-id>",
	Short: "pin an instance to the current directory via a .fablab-instance file",
	Args:  cobra.ExactArgs(1),
	Run:   pin,
}

var unpinCmd = &cobra.Command{
	Use:   "unpin",
	Short: "remove the nearest .fablab-instance pin file",
	Args:  cobra.ExactArgs(0),
	Run:   unpin,
}

func pin(_ *cobra.Command, args []string) {
	if err := model.PinInstance(args[0]); err != nil {
		logrus.Fatalf("error pinning instance (%v)", err)
	}
	fmt.Println("success")
}

func unpin(_ *cobra.Command, _ []string) {
	msg, err := model.UnpinInstance()
	if err != nil {
		logrus.Fatalf("error unpinning instance (%v)", err)
	}
	fmt.Println(msg)
}
