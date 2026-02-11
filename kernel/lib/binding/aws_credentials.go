package binding

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
)

var AwsCredentialsLoader model.BootstrapExtension = awsCredentialsLoader{}

type awsCredentialsLoader struct {
}

func (l awsCredentialsLoader) Bootstrap(m *model.Model) error {
	if m.GetScope().HasVariable("credentials.aws.access_key") &&
		m.GetScope().HasVariable("credentials.aws.secret_key") {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return errors.Errorf("couldn't load AWS config: %v", err)
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return errors.Errorf("couldn't load AWS credentials: %v", err)
	}

	m.PutVariable("credentials.aws.access_key", creds.AccessKeyID)
	m.PutVariable("credentials.aws.secret_key", creds.SecretAccessKey)

	return nil
}
