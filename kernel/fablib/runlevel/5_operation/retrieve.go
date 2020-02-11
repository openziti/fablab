package operation

import (
	"fmt"
	"github.com/netfoundry/fablab/kernel/fablib"
	"github.com/netfoundry/fablab/kernel/model"
	"os"
	"strings"
)

func Retrieve(region, host, path, extension string) model.OperatingStage {
	return &retrieve{
		region:    region,
		host:      host,
		path:      path,
		extension: extension,
	}
}

func (self *retrieve) Operate(m *model.Model, run string) error {
	hosts := m.GetHosts(self.region, self.host)
	if len(hosts) == 1 {
		ssh := fablib.NewSshConfigFactoryImpl(m, hosts[0].PublicIp)

		if files, err := fablib.RemoteFileList(ssh, self.path); err == nil {
			paths := make([]string, 0)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), self.extension) {
					paths = append(paths, file.Name())
				}
			}
			forensicsPath := model.AllocateForensicScenario(run, self.region)
			if err := os.MkdirAll(forensicsPath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating forensics root [%s] (%w)", forensicsPath, err)
			}
			if err := fablib.RetrieveRemoteFiles(ssh, forensicsPath, paths...); err != nil {
				return fmt.Errorf("error retrieving remote files (%w)", err)
			}
			if err := fablib.DeleteRemoteFiles(ssh, paths...); err != nil {
				return fmt.Errorf("error deleting remote files (%w)", err)
			}

		} else {
			return fmt.Errorf("error listing remote directory (%w)", err)
		}

	} else {
		return fmt.Errorf("found [%d] hosts", len(hosts))
	}

	return nil
}

type retrieve struct {
	region    string
	host      string
	path      string
	extension string
}
