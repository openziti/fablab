package subcmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	modelListCmd.AddCommand(listHostsCmd)
	modelCmd.AddCommand(modelListCmd)
	RootCmd.AddCommand(modelCmd)
}

var modelCmd = &cobra.Command{
	Use:     "model",
	Aliases: []string{"mod", "m"},
	Short:   "work with the model",
}

var modelListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list model entities",
}

var listHostsCmd = &cobra.Command{
	Use:   "hosts <spec?>",
	Short: "list hosts",
	Args:  cobra.MaximumNArgs(2),
	Run:   listHosts,
}

func listHosts(cmd *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	label := model.GetLabel()
	if label == nil {
		logrus.Fatalf("no label for instance [%s]", model.ActiveInstancePath())
	} else {
		m, found := model.GetModel(label.Model)
		if !found {
			logrus.Fatalf("no such model [%s]", label.Model)
		}

		if !m.IsBound() {
			logrus.Fatalf("model not bound")
		}

		hostSpec := "*"

		if len(args) > 0 {
			hostSpec = args[0]
		}

		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.AppendHeader(table.Row{"#", "ID", "Public IP", "Private IP", "Components", "Region", "Tags"})

		count := 0
		for _, host := range m.SelectHosts(hostSpec) {
			var components []string
			for component := range host.Components {
				components = append(components, component)
			}
			t.AppendRow(table.Row{count + 1, host.GetId(), host.PublicIp, host.PrivateIp,
				strings.Join(components, ","), host.GetRegion().Region,
				strings.Join(host.Tags, ",")})
			count++
		}

		if _, err := fmt.Fprintln(cmd.OutOrStdout(), t.Render()); err != nil {
			panic(err)
		}
	}
}
