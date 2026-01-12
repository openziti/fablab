// Package aws provides AWS-specific configuration types and utilities for Fablab models.
// This includes security group definitions, EC2 volume configurations, and network rules.
package aws

import (
	"bytes"
	"fmt"
	"strings"
)

// Env represents an environment that can mangle names for AWS resource uniqueness.
// Implementations typically prefix names with environment identifiers.
type Env interface {
	MangleName(name string) string
}

// DefaultSecurityGroup is a pre-configured security group with standard default rules.
// It includes SSH ingress on port 22 and unrestricted egress traffic.
var DefaultSecurityGroup = &SecurityGroup{
	Id:                  "default",
	ExcludeDefaultRules: false,
}

// SecurityGroup defines an AWS security group with network access rules.
// Security groups control inbound and outbound traffic for EC2 instances.
type SecurityGroup struct {
	// Id is the unique identifier for this security group
	Id string
	// Rules defines the network access rules for this security group
	Rules []*NetworkRule
	// env is the environment context used for name mangling
	env Env
	// ExcludeDefaultRules when true prevents automatic addition of SSH and egress rules
	ExcludeDefaultRules bool
}

// init initializes the security group with default rules and validates all rule configurations.
// If ExcludeDefaultRules is false, it adds SSH ingress (port 22) and unrestricted egress rules.
// Returns an error if any rule has an invalid direction, protocol, or missing CIDR blocks.
func (self *SecurityGroup) init(id string, env Env) error {
	self.Id = id
	self.env = env

	if !self.ExcludeDefaultRules {
		self.Rules = append(self.Rules, &NetworkRule{
			Direction: Ingress,
			Port:      22,
			Protocol:  "tcp",
		})
		self.Rules = append(self.Rules, &NetworkRule{
			Direction: Egress,
			Port:      0,
			Protocol:  "-1",
		})
	}

	for _, rule := range self.Rules {
		if rule.Direction != Ingress && rule.Direction != Egress {
			return fmt.Errorf("invalid rule direction: %v", rule.Direction)
		}

		if rule.Protocol != "udp" && rule.Protocol != "tcp" && rule.Protocol != "-1" {
			return fmt.Errorf("invalid rule protocol: %v", rule.Protocol)
		}

		if len(rule.CidrBlocks) == 0 {
			rule.CidrBlocks = append(rule.CidrBlocks, "0.0.0.0/0")
		}
	}

	return nil
}

// Name returns the environment-mangled name for this security group.
// This ensures uniqueness across different environments.
func (self *SecurityGroup) Name() string {
	return self.env.MangleName(self.Id)
}

// RuleDirection specifies whether a network rule applies to incoming or outgoing traffic.
type RuleDirection string

const (
	// Ingress represents inbound traffic rules
	Ingress RuleDirection = "ingress"
	// Egress represents outbound traffic rules
	Egress RuleDirection = "egress"
)

// SecurityGroups is a map of security group IDs to SecurityGroup instances.

type SecurityGroups map[string]*SecurityGroup

// Model represents the AWS-specific configuration within a Fablab model.
type Model struct {
	// SecurityGroups defines the security groups available for use by hosts and components
	SecurityGroups map[string]*SecurityGroup
}

// Init initializes all security groups in the model with the given environment context.
// Returns an error if any security group fails to initialize.
func (self *Model) Init(env Env) error {
	for k, sg := range self.SecurityGroups {
		if err := sg.init(k, env); err != nil {
			return err
		}
	}

	return nil
}

// NetworkRule defines a single network access rule for a security group.
type NetworkRule struct {
	// Direction specifies whether this rule applies to ingress or egress traffic
	Direction RuleDirection
	// Port specifies the network port (0 for all ports when Protocol is "-1")
	Port uint16
	// Protocol must be "tcp", "udp", or "-1" (all protocols)
	Protocol string
	// CidrBlocks defines the IP ranges this rule applies to (defaults to "0.0.0.0/0" if empty)
	CidrBlocks []string
}

// CidrBlockList returns a Terraform-formatted array string of CIDR blocks.
// CIDR blocks containing "/" are quoted as literals, others are treated as Terraform variables.
func (self *NetworkRule) CidrBlockList() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for idx, cidr := range self.CidrBlocks {
		if idx > 0 {
			buf.WriteString(",")
		}
		// if it contains a slash, assume it's a cidr, if not assume it's a terraform variable
		if strings.Contains(cidr, "/") {
			buf.WriteString(`"` + cidr + `"`)
		} else {
			buf.WriteString(cidr)
		}
	}
	buf.WriteString("]")

	return buf.String()
}

// EC2Volume defines the configuration for an EBS volume attached to an EC2 instance.
type EC2Volume struct {
	// Type specifies the EBS volume type (e.g., "gp3", "gp2", "io1", "io2")
	Type string
	// SizeGB is the volume size in gigabytes
	SizeGB uint32
	// IOPS specifies the provisioned IOPS (only for io1/io2 volume types)
	IOPS uint32
}

// EC2Host defines AWS-specific configuration for a Fablab host.
type EC2Host struct {
	// Volume specifies the EBS volume configuration
	Volume EC2Volume
	// SecurityGroup optionally specifies a custom security group ID for this host
	SecurityGroup string
}

// Component defines AWS-specific configuration for a Fablab component.
type Component struct {
	// SecurityGroup optionally specifies a custom security group ID for this component
	SecurityGroup string
}
