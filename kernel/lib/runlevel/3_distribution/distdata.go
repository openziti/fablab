package distribution

import (
	"fmt"
	"os"
	"strings"

	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func DistributeDataWithReplaceCallbacks(hostSpec, data, dest string, filemode os.FileMode, callbacks map[string]func(*model.Host) string) model.Stage {
	return &distDataWithReplaceCallbacks{
		hostSpec:  hostSpec,
		data:      data,
		dest:      dest,
		callbacks: callbacks,
		filemode:  filemode,
	}
}

func (df *distDataWithReplaceCallbacks) Execute(run model.Run) error {
	return run.GetModel().ForEachHost(df.hostSpec, 25, func(host *model.Host) error {
		return retryOnHost(host, func() error {
			dataRaw := df.data

			for k, v := range df.callbacks {
				dataRaw = strings.ReplaceAll(dataRaw, k, v(host))
			}

			if err := host.SendData([]byte(dataRaw), df.dest); err != nil {
				logrus.Errorf("[%s] unable to send data => %s", host.PublicIp, df.dest)
				return err
			}

			if _, err := host.ExecLogged(fmt.Sprintf("chmod %04o %s", df.filemode, df.dest)); err != nil {
				logrus.Errorf("[%s] unable to chmod %s", host.PublicIp, df.dest)
				return err
			}

			logrus.Infof("[%s] data => %s", host.PublicIp, df.dest)

			return nil
		})
	})
}

type distDataWithReplaceCallbacks struct {
	hostSpec  string
	data      string
	dest      string
	callbacks map[string]func(*model.Host) string
	filemode  os.FileMode
}

func DistributeData(hostSpec string, data []byte, dest string) model.Stage {
	return &distData{
		hostSpec: hostSpec,
		data:     data,
		dest:     dest,
	}
}

func (df *distData) Execute(run model.Run) error {
	return run.GetModel().ForEachHost(df.hostSpec, 25, func(host *model.Host) error {
		return retryOnHost(host, func() error {
			if err := host.SendData(df.data, df.dest); err != nil {
				logrus.Errorf("[%s] unable to send data => %s", host.PublicIp, df.dest)
				return err
			}

			if _, err := host.ExecLogged(fmt.Sprintf("chmod 0644 %s", df.dest)); err != nil {
				logrus.Errorf("[%s] unable to chmod %s", host.PublicIp, df.dest)
				return err
			}

			logrus.Infof("[%s] data => %s", host.PublicIp, df.dest)

			return nil
		})
	})
}

type distData struct {
	hostSpec string
	data     []byte
	dest     string
}
