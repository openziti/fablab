package model

import (
	"fmt"
	"reflect"
)

func (m *Model) Dump() *ModelDump {
	return &ModelDump{
		Scope: dumpScope(m.Scope),
	}
}

func dumpScope(s Scope) *ScopeDump {
	return &ScopeDump{
		Variables: dumpVariables(s.Variables),
	}
}

func dumpVariables(vs Variables) map[string]interface{} {
	dump := make(map[string]interface{})
	for k, v := range vs {
		kk := fmt.Sprintf("%v", k)
		if vvs, ok := v.(Variables); ok {
			dump[kk] = dumpVariables(vvs)

		} else if vv, ok := v.(*Variable); ok {
			dump[kk] = dumpVariable(vv)

		} else {
			dump[kk] = reflect.TypeOf(v).String()
		}
	}
	return dump
}

func dumpVariable(v *Variable) *VariableDump {
	dump := &VariableDump{
		Description: v.Description,
		Required: v.Required,
		Scoped: v.Scoped,
		GlobalFallback: v.GlobalFallback,
		Bound: v.bound,
	}
	if v.Default != nil {
		dump.Default = fmt.Sprintf("%v", v.Default)
	}
	if v.Binder != nil {
		dump.Binder = fmt.Sprintf("%p", v.Binder)
	}
	if v.Value != nil {
		dump.Value = fmt.Sprintf("%v", v.Value)
	}
	return dump
}

type ModelDump struct {
	Scope *ScopeDump `json:"scope"`
}

type ScopeDump struct {
	Variables map[string]interface{} `json:"variables"`
}

type VariableDump struct {
	Description    string `json:"description,omitempty"`
	Default        string `json:"default,omitempty"`
	Required       bool   `json:"required"`
	Scoped         bool   `json:"scoped"`
	GlobalFallback bool   `json:"global_fallback"`
	Binder         string `json:"binder,omitempty"`
	Value          string `json:"value,omitempty"`
	Bound          bool   `json:"bound"`
}
