/*
	Copyright 2019 Netfoundry, Inc.

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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParentBase(t *testing.T) {
	m := &Model{
		Parent: parentTestModel(),
	}
	assert.Nil(t, m.Merge(m.Parent))

	var found bool
	var value interface{}

	value, found = m.GetVariable("a")
	assert.True(t, found)
	assert.Equal(t, "oh, wow!", value)

	assert.Equal(t, 1, len(m.Regions))
	value, found = m.Regions["base"]
	assert.True(t, found)

	baseRegion := value.(*Region)
	assert.Equal(t, "us-east-1", baseRegion.Id)
	assert.Equal(t, "us-east-1a", baseRegion.Az)

	assert.Equal(t, 1, len(baseRegion.Hosts))
	value, found = baseRegion.Hosts["a"]
	assert.True(t, found)

	aHost := value.(*Host)
	assert.Equal(t, 1, len(aHost.Tags))
	assert.Equal(t, "0", aHost.Tags[0])

	assert.Equal(t, 1, len(m.Factories))
	assert.Nil(t, m.Data["factory"])
	m.Factories[0].Build(m)
	assert.Equal(t, "base", m.Data["factory"])
}

func TestParentMerge(t *testing.T) {
	m := &Model{
		Parent: parentTestModel(),
		Scope: Scope{
			Variables: Variables{
				"b": &Variable{Default: "hello!"},
			},
		},
		Regions: Regions{
			"base": {
				Az: "us-east-1b",
				Hosts: Hosts{
					"a": {
						Scope: Scope{Tags: Tags{"1"}},
					},
				},
			},
		},
		Factories: []Factory{
			&factory{name: "merge"},
		},
	}
	assert.Nil(t, m.Merge(m.Parent))

	var found bool
	var value interface{}

	value, found = m.GetVariable("a")
	assert.True(t, found)
	assert.Equal(t, "oh, wow!", value)
	value, found = m.GetVariable("b")
	assert.True(t, found)
	assert.Equal(t, "hello!", value)

	assert.Equal(t, 1, len(m.Regions))
	value, found = m.Regions["base"]
	assert.True(t, found)

	baseRegion := value.(*Region)
	assert.Equal(t, "us-east-1", baseRegion.Id)
	assert.Equal(t, "us-east-1b", baseRegion.Az)

	assert.Equal(t, 1, len(baseRegion.Hosts))
	value, found = baseRegion.Hosts["a"]
	assert.True(t, found)

	aHost := value.(*Host)
	assert.Equal(t, 2, len(aHost.Tags))
	assert.Equal(t, "0", aHost.Tags[0])
	assert.Equal(t, "1", aHost.Tags[1])

	assert.Equal(t, 2, len(m.Factories))
	assert.Nil(t, m.Data["factory"])
	m.Factories[0].Build(m)
	assert.Equal(t, "base", m.Data["factory"])
	m.Factories[1].Build(m)
	assert.Equal(t, "merge", m.Data["factory"])
}

func parentTestModel() *Model {
	return &Model {
		Scope: Scope{
			Variables: Variables{
				"a": &Variable{Default: "oh, wow!"},
			},
		},
		Regions: Regions{
			"base": {
				Id: "us-east-1",
				Az: "us-east-1a",
				Hosts: Hosts{
					"a": {
						Scope: Scope{Tags: Tags{"0"}},
					},
				},
			},
		},
		Factories: []Factory{
			&factory{name: "base"},
		},
	}
}

type factory struct{
	name string
}

func (f *factory) Build(m *Model) error {
	if m.Data == nil {
		m.Data = make(Data)
	}
	m.Data["factory"] = f.name
	return nil
}