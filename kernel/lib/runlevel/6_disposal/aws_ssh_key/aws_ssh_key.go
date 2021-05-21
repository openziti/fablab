package aws_ssh_key

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func Dispose() model.DisposalStage {
	return &awsKeyManager{}
}

type awsKeyManager struct {
}

func (stage awsKeyManager) Dispose(run model.Run) error {
	m := run.GetModel()
	if managedKey, found := m.GetBoolVariable("credentials", "aws", "managed_key"); !found || !managedKey {
		return nil
	}
	keyName, found := m.GetStringVariable("credentials", "aws", "ssh_key_name")
	if !found {
		keyName = m.MustStringVariable("environment")
	}

	awsAccessKey := m.MustStringVariable("credentials", "aws", "access_key")
	awsSecretKey := m.MustStringVariable("credentials", "aws", "secret_key")

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

func getPublicKey(privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.Errorf("failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKeyDer, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	pubKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   publicKeyDer,
	}

	return pem.EncodeToMemory(&pubKeyBlock), nil
}
