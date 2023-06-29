package aws_ssh_key

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	awsCreds := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")

	for _, region := range m.Regions {
		logrus.Infof("removing key '%v' from region %v", keyName, region.Region)
		awsConfig := &aws.Config{
			Credentials: awsCreds,
			Region:      &region.Region,
		}
		awsSession, err := session.NewSession(awsConfig)
		if err != nil {
			return err
		}
		ec2Client := ec2.New(awsSession)
		input := &ec2.DeleteKeyPairInput{KeyName: &keyName}
		if _, err = ec2Client.DeleteKeyPair(input); err != nil {
			return err
		}
	}

	return nil
}
