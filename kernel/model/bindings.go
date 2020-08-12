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
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (m *Model) BindBindings(bindings Bindings) error {
	errors := make([]error, 0)
	m.IterateScopes(func(i interface{}, path ...string) {
		var v Variables
		if m, ok := i.(*Model); ok {
			v = m.Variables
		} else if r, ok := i.(*Region); ok {
			v = r.Variables
		} else if h, ok := i.(*Host); ok {
			v = h.Variables
		} else if c, ok := i.(*Component); ok {
			v = c.Variables
		}
		err := bindings.bindMap(i, v, path, nil)
		if err != nil {
			errors = append(errors, err)
		}
	})
	if len(errors) > 0 {
		msg := "errors {"
		for i := 0; i < len(errors); i++ {
			msg += fmt.Sprintf("[%d]: %s", i, errors[i])
		}
		msg += "}"
		return fmt.Errorf("cannot bind: %s", msg)
	}
	return nil
}

func (bindings Bindings) Has(name ...string) bool {
	_, found := bindings.Get(name...)
	return found
}

func (bindings Bindings) Put(value interface{}, rootKey string, rest ...string) {
	if len(rest) == 0 {
		bindings[rootKey] = value
		return
	}

	var lowerMap Bindings
	if value, found := bindings[rootKey]; found {
		lowerMap, _ = value.(Bindings)
	}
	if lowerMap == nil {
		lowerMap = Bindings{}
		bindings[rootKey] = lowerMap
	}
	lowerMap.Put(value, rest[0], rest[1:]...)
}

func (bindings Bindings) Must(name ...string) interface{} {
	value, found := bindings.Get(name...)
	if !found {
		logrus.Fatalf("binding [%v] not found", name)
	}
	return value
}

func (bindings Bindings) Get(name ...string) (interface{}, bool) {
	if len(name) < 1 {
		return nil, false
	}

	inputMap := bindings
	for i := 0; i < (len(name) - 1); i++ {
		key := name[i]
		if value, found := inputMap[key]; found {
			lowerMap, ok := value.(Bindings)
			if !ok {
				return nil, false
			}
			inputMap = lowerMap
		} else {
			return nil, false
		}
	}

	if value, found := inputMap[name[len(name)-1]]; found {
		return value, true
	}
	return nil, false
}

func (bindings Bindings) GetString(name ...string) (string, bool) {
	val, found := bindings.Get(name...)
	if found {
		result, ok := val.(string)
		return result, ok
	}
	return "", found
}

func (bindings Bindings) GetBool(name ...string) (bool, bool) {
	val, found := bindings.Get(name...)
	if found {
		result, ok := val.(bool)
		return result, ok
	}
	return false, found
}

func (bindings Bindings) bindMap(i interface{}, variables Variables, scopePath []string, parent []string) error {
	for k, v := range variables {
		variable, ok := v.(*Variable)
		if ok {
			path := append(parent, k.(string))
			if variable.Required {
				if variable.Scoped {
					variable.Value, ok = bindings.Get(append(scopePath, path...)...)
					if !ok {
						if variable.GlobalFallback {
							variable.Value, ok = bindings.Get(path...)
							if !ok {
								return fmt.Errorf("error binding scoped, required variable %v at %v", path, scopePath)
							}

						} else {
							return fmt.Errorf("error binding scoped, required variable %v at %v", path, scopePath)
						}
					}
				} else {
					variable.Value, ok = bindings.Get(path...)
					if !ok {
						return fmt.Errorf("error binding required variable %v at %v", path, scopePath)
					}
				}
				variable.bound = true
				if variable.Binder != nil {
					variable.Binder(variable, i, scopePath...)
				}
				logrus.Debugf("bound required variable %v = [%s] at %v", path, variable.Value, scopePath)

			} else {
				var fullPath []string
				if variable.Scoped {
					fullPath = append(scopePath, path...)
				} else {
					fullPath = path
				}

				value, found := bindings.Get(fullPath...)
				if found {
					variable.Value = value
					variable.bound = true
					if variable.Binder != nil {
						variable.Binder(variable, i, scopePath...)
					}
					logrus.Debugf("bound optional variable %v = [%s] at %v", path, variable.Value, scopePath)

				} else if variable.GlobalFallback && variable.Scoped {
					value, found := bindings.Get(path...)
					if found {
						variable.Value = value
						variable.bound = true
						if variable.Binder != nil {
							variable.Binder(variable, i, scopePath...)
						}
						logrus.Debugf("bound optional variable to unscoped global %v = [%s] at %v", path, variable.Value, scopePath)
					}
				}

				if variable.Default != nil && !variable.bound {
					variable.Value = variable.Default
					variable.bound = true
					if variable.Binder != nil {
						variable.Binder(variable, i, scopePath...)
					}
					logrus.Debugf("bound optional variable to default %v = [%s] at %v", path, variable.Value, scopePath)
				}
			}
		} else {
			nextmap, ok := v.(Variables)
			if ok {
				err := bindings.bindMap(i, nextmap, scopePath, append(parent, k.(string)))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type Bindings map[interface{}]interface{}

func loadBindings() error {
	var data []byte
	var err error
	data, err = ioutil.ReadFile(bindingsYml())
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("no bindings [%s]", bindingsYml())
		} else {
			return fmt.Errorf("error reading bindings [%s] (%w)", bindingsYml(), err)
		}
	}

	bindings = make(map[interface{}]interface{})
	if err := yaml.Unmarshal(data, &bindings); err != nil {
		return fmt.Errorf("error unmarshalling bindings [%s] (%w)", bindingsYml(), err)
	}

	return nil
}

func bindingsYml() string {
	return filepath.Join(configRoot(), "bindings.yml")
}
