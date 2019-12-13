package subcmd

import (
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/netfoundry/fablab/zitilab/console"
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
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	l := model.GetLabel()
	if l == nil {
		logrus.Fatalf("no label for instance [%s]", model.ActiveInstancePath())
	}

	if l != nil {
		_, found := model.GetModel(l.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", l.Model)
		}

		server := console.NewServer()
		go server.Listen()

		http.Handle("/", http.FileServer(http.Dir("zitilab/console/webroot")))
		logrus.Fatal(http.ListenAndServe(":8080", nil))

	} else {
		logrus.Fatalf("no label for run")
	}
}
