package aws_ssh_key

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var KeyManager = awsKeyManager{}

func Express() model.Stage {
	return KeyManager
}

type awsKeyManager struct{}

func (stage awsKeyManager) Bootstrap(m *model.Model) error {
	if !m.HasVariable("credentials.aws.ssh_key_name") {
		environment := m.MustStringVariable("environment")
		instanceId := model.ActiveInstanceId()
		keyName := fmt.Sprintf("%v-%v", environment, instanceId)
		m.PutVariable("credentials.aws.ssh_key_name", keyName)
	}

	if m.HasVariable("credentials.ssh.key_path") {
		return nil
	}

	keyPath := path.Join(model.BuildPath(), "ssh_private_key.pem")
	m.PutVariable("credentials.ssh.key_path", keyPath)

	return nil
}

func (stage awsKeyManager) Execute(run model.Run) error {
	m := run.GetModel()

	if managedKey, found := m.GetBoolVariable("credentials.aws.managed_key"); !found || !managedKey {
		if !found {
			logrus.Info("credentials.aws.managed_key setting not found. skipping managed key setup")
		} else if !managedKey {
			logrus.Info("credentials.aws.managed_key setting set to false. skipping managed key setup")
		}

		return nil
	}

	logrus.Info("beginning managed key setup")
	keyName, found := m.GetStringVariable("credentials.aws.ssh_key_name")
	if !found {
		keyName = m.MustStringVariable("environment")
	}

	awsAccessKey := m.MustStringVariable("credentials.aws.access_key")
	awsSecretKey := m.MustStringVariable("credentials.aws.secret_key")

	awsCreds := credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")

	var privateKey []byte
	var publicKey []byte

	ctx := context.Background()
	keyPath := m.MustStringVariable("credentials.ssh.key_path")
	logrus.Infof("checking for  private key in %v", keyPath)
	var err error
	if privateKey, err = os.ReadFile(keyPath); err == nil {
		logrus.Infof("loaded private key from %v... deriving public key", keyPath)
		publicKey, err = getPublicKey(privateKey)
		if err != nil {
			return err
		}
	} else {
		logrus.Infof("failed to load private key from %v (%v), generating new key", keyPath, err)

		for _, region := range m.Regions {
			cfg := aws.Config{
				Credentials: awsCreds,
				Region:      region.Region,
			}
			ec2Client := ec2.NewFromConfig(cfg)

			_, err = ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
				KeyNames: []string{keyName},
			})

			if err == nil {
				logrus.Infof("removing key %v from region %v, as we don't have the private key anymore", keyPath, region.Region)
				if _, err = ec2Client.DeleteKeyPair(ctx, &ec2.DeleteKeyPairInput{KeyName: &keyName}); err != nil {
					return errors.Wrapf(err, "failed to remove private key %v from region %v", keyName, region.Region)
				}
			}
		}
	}

	for _, region := range m.Regions {
		cfg := aws.Config{
			Credentials: awsCreds,
			Region:      region.Region,
		}
		ec2Client := ec2.NewFromConfig(cfg)

		_, err = ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
			KeyNames: []string{keyName},
		})

		if err == nil {
			logrus.Infof("key pair %v already exists in region %v. skipping create/import", keyName, region.Region)
			continue
		}

		if publicKey == nil {
			logrus.Infof("creating key '%v' in region %v", keyName, region.Region)
			keyPairInput := &ec2.CreateKeyPairInput{KeyName: &keyName}
			output, err := ec2Client.CreateKeyPair(ctx, keyPairInput)
			if err != nil {
				return err
			}
			privateKey = []byte(*output.KeyMaterial)
			publicKey, err = getPublicKey(privateKey)
			if err != nil {
				return err
			}
			keyPath := m.MustStringVariable("credentials.ssh.key_path")
			logrus.Infof("saving private key '%v' to %v", keyName, keyPath)
			if err = os.WriteFile(keyPath, privateKey, 0600); err != nil {
				return err
			}
		} else {
			logrus.Infof("importing key '%v' in region %v", keyName, region.Region)
			keyPairInput := &ec2.ImportKeyPairInput{
				KeyName:           &keyName,
				PublicKeyMaterial: publicKey,
			}
			if _, err := ec2Client.ImportKeyPair(ctx, keyPairInput); err != nil {
				return err
			}
		}
	}
	return nil
}

func getPublicKey(privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block != nil && block.Type == "RSA PRIVATE KEY" {
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		publicKey, err := ssh.NewPublicKey(&key.PublicKey)
		if err != nil {
			return nil, err
		}
		return ssh.MarshalAuthorizedKey(publicKey), nil
	}

	if block != nil && block.Type == "OPENSSH PRIVATE KEY" {
		key, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			pfxlog.Logger().Errorf("error parsing PK (%v)", err)
			return nil, err
		}
		return ssh.MarshalAuthorizedKey(key.PublicKey()), nil
	}

	return nil, errors.Errorf("failed to decode PEM block containing public key")
}
