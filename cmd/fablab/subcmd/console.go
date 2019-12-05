package subcmd

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/console"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

func init() {
	RootCmd.AddCommand(consoleCmd)
}

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "local web console",
	Args:  cobra.ExactArgs(0),
	Run:   doConsole,
}

func doConsole(_ *cobra.Command, _ []string) {
	if err := kernel.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	l := kernel.GetLabel()
	if l == nil {
		logrus.Fatalf("no label for instance [%s]", kernel.ActiveInstancePath())
	}

	if l != nil {
		_, found := kernel.GetModel(l.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", l.Model)
		}

		server := console.NewServer()
		go server.Listen()

		http.Handle("/", http.FileServer(http.Dir("console/webroot")))
		logrus.Fatal(http.ListenAndServe(":8080", nil))

	} else {
		logrus.Fatalf("no label for run")
	}
}
