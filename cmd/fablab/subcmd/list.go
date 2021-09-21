package subcmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

func init() {
	listCmd.AddCommand(listInstancesCmd)
	listCmd.AddCommand(listHostsCmd)
	listCmd.AddCommand(listComponentsCmd)
	RootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list model entities",
}

var listInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "list instances",
	Args:  cobra.ExactArgs(0),
	Run:   listInstances,
}

func listInstances(_ *cobra.Command, _ []string) {
	if err := model.BootstrapInstance(); err != nil {
		logrus.Fatalf("unable to bootstrap config (%v)", err)
	}

	activeInstanceId := model.ActiveInstanceId()

	cfg := model.GetConfig()

	var instanceIds []string
	for k := range cfg.Instances {
		instanceIds = append(instanceIds, k)
	}

	sort.Strings(instanceIds)

	fmt.Println()
	fmt.Printf("[%d] instances:\n\n", len(instanceIds))
	for _, instanceId := range instanceIds {
		idLabel := instanceId
		if instanceId == activeInstanceId {
			idLabel += "*"
		}
		instanceConfig := cfg.Instances[instanceId]
		if l, err := instanceConfig.LoadLabel(); err == nil {
			fmt.Printf("%-12s %-24s [%s]\n", idLabel, l.Model, l.State)
		} else {
			fmt.Printf("%-12s %s\n", idLabel, err)
		}
	}
	if len(instanceIds) > 0 {
		fmt.Println()
	}
}

var listHostsCmd = &cobra.Command{
	Use:   "hosts <spec?>",
	Short: "list hosts",
	Args:  cobra.MaximumNArgs(1),
	Run:   listHosts,
}

var listComponentsCmd = &cobra.Command{
	Use:     "components <spec?>",
	Aliases: []string{"comp"},
	Short:   "list components",
	Args:    cobra.MaximumNArgs(1),
	Run:     listComponents,
}

func listHosts(cmd *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()

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

func listComponents(cmd *cobra.Command, args []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()
	componentSpec := "*"

	if len(args) > 0 {
		componentSpec = args[0]
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"#", "ID", "Host", "Region", "Tags"})

	count := 0
	for _, c := range m.SelectComponents(componentSpec) {
		t.AppendRow(table.Row{count + 1, c.GetId(), c.GetHost().GetId(), c.GetRegion().GetId(), strings.Join(c.Tags, ",")})
		count++
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), t.Render()); err != nil {
		panic(err)
	}
}
