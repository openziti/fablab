package aws_ssh_key

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Dispose() model.Stage {
	return &awsKeyManager{}
}

type awsKeyManager struct {
}

func (stage awsKeyManager) Execute(run model.Run) error {
	m := run.GetModel()
	if managedKey, found := m.GetBoolVariable("credentials.aws.managed_key"); !found || !managedKey {
		return nil
	}
	keyName, found := m.GetStringVariable("credentials.aws.ssh_key_name")
	if !found {
		keyName = m.MustStringVariable("environment")
	}

	awsAccessKey := m.MustStringVariable("credentials.aws.access_key")
	awsSecretKey := m.MustStringVariable("credentials.aws.secret_key")

	awsCreds := credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")

	ctx := context.Background()
	for _, region := range m.Regions {
		logrus.Infof("removing key '%v' from region %v", keyName, region.Region)
		cfg := aws.Config{
			Credentials: awsCreds,
			Region:      region.Region,
		}
		ec2Client := ec2.NewFromConfig(cfg)
		input := &ec2.DeleteKeyPairInput{KeyName: &keyName}
		if _, err := ec2Client.DeleteKeyPair(ctx, input); err != nil {
			return err
		}
	}

	return nil
}
