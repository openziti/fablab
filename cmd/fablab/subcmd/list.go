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
	listCmd.AddCommand(newListHostsCmd())
	listCmd.AddCommand(newListComponentsCmd())
	listCmd.AddCommand(listActionsCmd)
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

func newListHostsCmd() *cobra.Command {
	action := &listHostsAction{}

	var cmd = &cobra.Command{
		Use:   "hosts <spec?>",
		Short: "list hosts",
		Args:  cobra.MaximumNArgs(1),
		Run:   action.execute,
	}

	cmd.Flags().BoolVarP(&action.componentDetail, "detail", "d", false, "show extra component detail")

	return cmd
}

func newListComponentsCmd() *cobra.Command {
	action := listComponentsAction{}

	var cmd = &cobra.Command{
		Use:     "components <spec?>",
		Aliases: []string{"comp"},
		Short:   "list components",
		Args:    cobra.MaximumNArgs(1),
		Run:     action.execute,
	}

	return cmd
}

var listActionsCmd = &cobra.Command{
	Use:   "actions",
	Short: "list actions",
	Args:  cobra.MaximumNArgs(1),
	Run:   listActions,
}

func listInstances(_ *cobra.Command, _ []string) {
	cfg := model.GetConfig()
	activeInstanceId := cfg.Default

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

type listHostsAction struct {
	componentDetail bool
}

func (self *listHostsAction) execute(cmd *cobra.Command, args []string) {
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
	t.AppendHeader(table.Row{"#", "ID", "Public IP", "Private IP", "Components", "Region", "InstanceType", "Tags"})

	count := 0
	for _, host := range m.SelectHosts(hostSpec) {
		components := &strings.Builder{}
		first := true

		if self.componentDetail {
			for _, component := range host.Components {
				cType := "none"
				if component.Type != nil {
					cType = component.Type.Label()
				}

				if !first {
					components.WriteString("\n")
				}
				components.WriteString(fmt.Sprintf("%s: %s", component.Id, cType))
				first = false
			}
		} else {
			summary := map[string]int{}
			for _, component := range host.Components {
				cType := "none"
				if component.Type != nil {
					cType = component.Type.Label()
				}
				summary[cType]++
			}
			for cType, count := range summary {
				if !first {
					components.WriteString("\n")
				}
				components.WriteString(fmt.Sprintf("%s: %4d", cType, count))
				first = false
			}
		}

		t.AppendRow(table.Row{count + 1, host.GetId(), host.PublicIp, host.PrivateIp,
			components.String(), host.GetRegion().Region, host.InstanceType,
			strings.Join(host.Tags, ",")})
		count++
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), t.Render()); err != nil {
		panic(err)
	}
}

type listComponentsAction struct{}

func (self *listComponentsAction) execute(cmd *cobra.Command, args []string) {
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
	t.AppendHeader(table.Row{"#", "ID", "Type", "Host", "Region", "Version", "Tags"})

	count := 0
	for _, c := range m.SelectComponents(componentSpec) {
		componentType := "none"
		componentVersion := ""
		if c.Type != nil {
			componentType = c.Type.Label()
			componentVersion = c.Type.GetVersion()
		}
		t.AppendRow(table.Row{count + 1, c.GetId(), componentType, c.GetHost().GetId(), c.GetRegion().GetId(),
			componentVersion, strings.Join(c.Tags, ",")})
		count++
	}

	if _, err := fmt.Fprintln(cmd.OutOrStdout(), t.Render()); err != nil {
		panic(err)
	}
}

func listActions(*cobra.Command, []string) {
	if err := model.Bootstrap(); err != nil {
		logrus.Fatalf("unable to bootstrap (%s)", err)
	}

	m := model.GetModel()

	for _, action := range m.GetActions() {
		fmt.Println(action)
	}
}
