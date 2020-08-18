package binding

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
)

var AwsCredentialsLoader model.BootstrapExtension = awsCredentialsLoader{}

type awsCredentialsLoader struct {
}

func (l awsCredentialsLoader) Bootstrap(m *model.Model) error {
	bindings := model.GetBindings()
	if bindings.Has("credentials", "aws", "access_key") &&
		bindings.Has("credentials", "aws", "secret_key") {
		return nil
	}

	val, err := defaults.Get().Config.Credentials.Get()
	if err != nil {
		return errors.Errorf("couldn't load AWS credentials: %v", err)
	}

	bindings.Put(val.AccessKeyID, "credentials", "aws", "access_key")
	bindings.Put(val.SecretAccessKey, "credentials", "aws", "secret_key")

	return nil
}
