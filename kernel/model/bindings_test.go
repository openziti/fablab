/*
	Copyright 2019 NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBindBindingsRequiredToModel(t *testing.T) {
	bValue := "b-value"

	binder := false
	m := &Model{
		Scope: Scope{
			Variables: Variables{
				"a": Variables{
					"b": &Variable{
						Required: true,
						Binder: func(v *Variable, i interface{}, path ...string) {
							_, ok := i.(*Model)
							assert.True(t, ok)
							assert.Equal(t, v.Value, bValue)
							binder = true
						},
					},
				},
			},
		},
	}

	bindings := Bindings{
		"a": Bindings{
			"b": bValue,
		},
		"c": "c-value",
	}
	err := m.BindBindings(bindings)
	assert.Nil(t, err)
	assert.True(t, binder)

	bindings = Bindings{
		"c": "c-value",
	}
	err = m.BindBindings(bindings)
	assert.NotNil(t, err)
	fmt.Printf("expected error: %s", err)
}

func TestBindBindingsScopedRequiredToComponent(t *testing.T) {
	binder := false
	configValue := "oh, wow!"
	configVariable := &Variable{
		Required: true,
		Binder: func(v *Variable, i interface{}, path ...string) {
			_, ok := i.(*Component)
			assert.True(t, ok)
			assert.Equal(t, v.Value, configValue)
			binder = true
		},
	}

	m := &Model{
		Regions: Regions{
			"region_0": &Region{
				Hosts: Hosts{
					"host_0": &Host{
						Components: Components{
							"component_0": &Component{
								Scope: Scope{
									Variables: Variables{
										"config": configVariable,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	bindings := Bindings{
		"config": configValue,
	}
	configVariable.Scoped = false
	err := m.BindBindings(bindings)
	assert.Nil(t, err)
	assert.True(t, binder)

	binder = false
	configVariable.Scoped = true
	configVariable.GlobalFallback = true
	err = m.BindBindings(bindings)
	assert.Nil(t, err)
	assert.True(t, binder)

	bindings = Bindings{
		"region_0": Bindings{
			"host_0": Bindings{
				"component_0": Bindings{
					"config": configValue,
				},
			},
		},
	}
	binder = false
	configVariable.Scoped = true
	configVariable.GlobalFallback = false
	err = m.BindBindings(bindings)
	assert.Nil(t, err)
	assert.True(t, binder)
}

func TestGetBinding(t *testing.T) {
	bindings := Bindings{
		"a": Bindings{
			"b": Bindings{
				"c": "hello",
			},
		},
		"d": "oh, wow!",
	}

	value, found := bindings.Get("a", "b", "c")
	assert.True(t, found)
	assert.Equal(t, "hello", value)

	value, found = bindings.Get("d")
	assert.True(t, found)
	assert.Equal(t, "oh, wow!", value)

	_, found = bindings.Get("e", "f", "d")
	assert.False(t, found)
}
