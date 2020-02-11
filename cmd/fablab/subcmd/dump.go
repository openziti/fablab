package subcmd

import (
	"encoding/json"
	"fmt"
	"github.com/netfoundry/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(dumpCmd)
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump the resolved model structure",
	Args:  cobra.ExactArgs(0),
	Run:   dump,
}

func dump(_ *cobra.Command, _ []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	l := model.GetLabel()
	if l == nil {
		logrus.Fatalf("no label for instance [%s]", model.ActiveInstancePath())
	}

	if l != nil {
		m, found := model.GetModel(l.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", l.Model)
		}

		if data, err := json.MarshalIndent(m.Dump(), "", "  "); err == nil {
			fmt.Println()
			fmt.Println(string(data))
		} else {
			logrus.Fatalf("error marshaling model dump (%w)", err)
		}

	} else {
		logrus.Fatalf("no label for run")
	}
}
