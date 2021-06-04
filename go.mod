module github.com/openziti/fablab

go 1.16

replace go.etcd.io/bbolt => github.com/openziti/bbolt v1.3.6-0.20210317142109-547da822475e

require (
	github.com/aws/aws-sdk-go v1.31.14
	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d
	github.com/jedib0t/go-pretty/v6 v6.0.4
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/michaelquigley/figlet v0.0.0-20191015203154-054d06db54b4
	github.com/michaelquigley/pfxlog v0.3.7
	github.com/natefinch/npipe v0.0.0-20160621034901-c1b8fa8bdcce
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/openziti/foundation v0.15.50
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v2 v2.4.0
)
