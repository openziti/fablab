package pki

import (
	"github.com/netfoundry/fablab/kernel"
	"github.com/netfoundry/fablab/kernel/lib"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func DotZiti() kernel.ConfigurationStage {
	return &dotZiti{}
}

func (d *dotZiti) Configure(m *kernel.Model) error {
	if err := generateCert("dotziti", "127.0.0.1"); err != nil {
		return fmt.Errorf("error generating cert for [dotziti] (%s)", err)
	}
	if err := generateLocalIdentities(m); err != nil {
		return fmt.Errorf("error generating local identities for [dotziti] (%s)", err)
	}
	if err := mergeLocalIdentities(); err != nil {
		return fmt.Errorf("error merging local identities for [dotziti] (%s)", err)
	}
	return nil
}

type dotZiti struct {
}

func generateLocalIdentities(m *kernel.Model) error {
	tPath := filepath.Join(kernel.ConfigSrc(), "local_identities.yml")
	tData, err := ioutil.ReadFile(tPath)
	if err != nil {
		return fmt.Errorf("error reading template [%s] (%s)", tPath, err)
	}

	t, err := template.New("config").Funcs(lib.TemplateFuncMap(m)).Parse(string(tData))
	if err != nil {
		return fmt.Errorf("error parsing template [%s] (%s)", tPath, err)
	}

	outputPath := filepath.Join(kernel.PkiBuild(), "local_identities.yml")
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating directories [%s] (%s)", outputPath, err)
	}

	outputF, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating config [%s] (%s)", outputPath, err)
	}
	defer func() { _ = outputF.Close() }()

	err = t.Execute(outputF, struct {
		RunPath string
	}{
		RunPath: kernel.ActiveInstancePath(),
	})
	if err != nil {
		return fmt.Errorf("error rendering template [%s] (%s)", outputPath, err)
	}

	logrus.Infof("config => [local_identities.yml]")

	return nil
}

func mergeLocalIdentities() error {
	home := os.Getenv("ZITI_HOME")

	var err error
	idPath := ""
	if home != "" {
		idPath = filepath.Join(home, "identities.yml")
	} else {
		home, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory (%s)", err)
		}
		idPath = filepath.Join(home, ".ziti/identities.yml")
	}

	var identities map[interface{}]interface{}
	_, err = os.Stat(idPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(idPath), os.ModePerm); err != nil {
				return fmt.Errorf("error making parent directory [%s] (%s)", filepath.Dir(idPath), err)
			}
			identities = make(map[interface{}]interface{})
		}
	} else {
		data, err := ioutil.ReadFile(idPath)
		if err != nil {
			return fmt.Errorf("unable to read existing identities [%s] (%s)", idPath, err)
		}

		err = yaml.Unmarshal(data, &identities)
		if err != nil {
			return fmt.Errorf("error unmarshaling existing identities [%s] (%s)", idPath, err)
		}
	}

	var localIdentities map[interface{}]interface{}
	localIdPath := filepath.Join(kernel.PkiBuild(), "local_identities.yml")
	data, err := ioutil.ReadFile(localIdPath)
	if err != nil {
		return fmt.Errorf("error reading local identities [%s] (%s)", localIdPath, err)
	}

	err = yaml.Unmarshal(data, &localIdentities)
	if err != nil {
		return fmt.Errorf("error unmarshalling local identities [%s] (%s)", localIdPath, err)
	}

	fablabIdentity, found := localIdentities["default"]
	if !found {
		return fmt.Errorf("no 'default' identity in local identities [%s] (%s)", localIdPath, err)
	}

	identities["fablab"] = fablabIdentity
	data, err = yaml.Marshal(identities)

	if err := ioutil.WriteFile(idPath, data, os.ModePerm); err != nil {
		return fmt.Errorf("error writing user identities [%s] (%s)", idPath, err)
	}

	return nil
}
