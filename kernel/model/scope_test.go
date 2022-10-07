/*
	Copyright NetFoundry Inc.

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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVariablesPut(t *testing.T) {
	v := Variables{
		"a": Variables{
			"b": "hello",
		},
	}
	value, found := v.Get([]string{"a", "b"})
	assert.True(t, found)
	assert.Equal(t, "hello", value)

	v.Put([]string{"a", "b"}, "oh, wow")
	value, found = v.Get([]string{"a", "b"})
	assert.True(t, found)
	assert.Equal(t, "oh, wow", value)
}

func TestVariableResolver(t *testing.T) {
	m := newTestModel()
	m.init()
	region := m.Regions["region1"]
	host := region.Hosts["host1"]
	component := host.Components["component1"]

	req := require.New(t)
	val, found := m.GetStringVariable("test.key")
	req.True(found)
	req.Equal("model.hello", val)

	val, found = region.GetStringVariable("test.key")
	req.True(found)
	req.Equal("region.hello", val)

	val, found = region.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.bye", val)

	val, found = host.GetStringVariable("test.key")
	req.True(found)
	req.Equal("host.hello", val)

	val, found = host.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.bye", val)

	val, found = component.GetStringVariable("test.key")
	req.True(found)
	req.Equal("hello", val)

	val, found = component.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.bye", val)
}

func TestVariableResolverBindingsOverride(t *testing.T) {
	defer func() {
		bindings = Variables{}
	}()

	bindings = Variables{
		"region1": Variables{
			"test": Variables{
				"key":  "region.override",
				"key2": "region.cascade",
			},
			"host1": Variables{
				"component1": Variables{
					"test": Variables{
						"key": "component.override",
					},
				},
			},
		},
		"other": Variables{
			"key": 77,
		},
	}

	m := newTestModel()
	m.VarConfig.ResolverLogger = func(resolver string, entity Entity, name string, result interface{}, found bool, msgAndArgs ...interface{}) {
		msg := ""
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(", ctx=%v", msgAndArgs[0])
			if len(msg) > 1 {
				msg = fmt.Sprintf(msg, msgAndArgs[1:]...)
			}
		}
		fmt.Printf("%v: %v[id=%v] key=%v result=%v, found=%v%v\n", resolver, entity.GetType(), entity.GetId(), name, result, found, msg)
	}
	m.init()
	region := m.Regions["region1"]
	host := region.Hosts["host1"]
	component := host.Components["component1"]

	req := require.New(t)
	fmt.Println("\nmodel:")
	val, found := m.GetStringVariable("test.key")
	req.True(found)
	req.Equal("model.hello", val)

	fmt.Println("\nregion:")
	val, found = region.GetStringVariable("test.key")
	req.True(found)
	req.Equal("region.override", val)

	fmt.Println("\nregion.2:")
	val, found = region.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.cascade", val)

	fmt.Println("\nhost:")
	val, found = host.GetStringVariable("test.key")
	req.True(found)
	req.Equal("host.hello", val)

	fmt.Println("\nhost.2:")
	val, found = host.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.cascade", val)

	fmt.Println("\ncomponent:")
	val, found = component.GetStringVariable("test.key")
	req.True(found)
	req.Equal("component.override", val)

	fmt.Println("\ncomponent.2:")
	val, found = component.GetStringVariable("test.key2")
	req.True(found)
	req.Equal("region.cascade", val)

	fmt.Println("\nother:")
	v, found := component.GetVariable("other.key")
	req.True(found)
	req.Equal(77, v)

}

func newTestModel() *Model {
	return &Model{
		Id: "test",
		Scope: Scope{
			Defaults: Variables{
				"test": Variables{
					"key": "model.hello",
				},
			},
		},
		Regions: Regions{
			"region1": {
				Scope: Scope{
					Defaults: Variables{
						"test": Variables{
							"key":  "region.hello",
							"key2": "region.bye",
						},
					},
				},
				Hosts: Hosts{
					"host1": {
						Scope: Scope{
							Defaults: Variables{
								"test": Variables{
									"key": "host.hello",
								},
							},
						},
						Components: Components{
							"component1": {
								Scope: Scope{
									Defaults: Variables{
										"test": Variables{
											"key": "hello",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
