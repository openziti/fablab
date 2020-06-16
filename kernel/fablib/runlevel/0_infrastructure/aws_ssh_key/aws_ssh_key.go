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
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"path"
)

var KeyManager = awsKeyManager{}

func Express() model.InfrastructureStage {
	return KeyManager
}

type awsKeyManager struct {
}

func (l awsKeyManager) Bootstrap(m *model.Model) error {
	bindings := model.GetBindings()

	if !bindings.Has("credentials", "aws", "ssh_key_name") {
		bindings.Put(bindings.Must("environment"), "credentials", "aws", "ssh_key_name")
	}

	if bindings.Has("credentials", "ssh", "key_path") {
		return nil
	}

	keyPath := path.Join(model.ActiveInstancePath(), "ssh_private_key.pem")

	bindings.Put(keyPath, "credentials", "ssh", "key_path")
	return nil
}

func (stage awsKeyManager) Express(m *model.Model, l *model.Label) error {
	bindings := model.GetBindings()
	if managedKey, found := bindings.GetBool("credentials", "aws", "managed_key"); !found || !managedKey {
		return nil
	}
	logrus.Info("beginning managed key setup")
	keyName, found := bindings.GetString("credentials", "aws", "ssh_key_name")
	if !found {
		keyName = m.MustStringVariable("environment")
	}

	awsAccessKey := m.MustStringVariable("credentials", "aws", "access_key")
	awsSecretKey := m.MustStringVariable("credentials", "aws", "secret_key")

	awsCreds := credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, "")

	var privateKey []byte
	var publicKey []byte

	for _, region := range m.Regions {
		awsConfig := &aws.Config{
			Credentials: awsCreds,
			Region:      &region.Id,
		}
		awsSession, err := session.NewSession(awsConfig)
		if err != nil {
			return err
		}
		ec2Client := ec2.New(awsSession)

		if publicKey == nil {
			logrus.Infof("creating key '%v' in region %v", keyName, region.Id)
			keyPairInput := &ec2.CreateKeyPairInput{KeyName: &keyName}
			output, err := ec2Client.CreateKeyPair(keyPairInput)
			if err != nil {
				return err
			}
			privateKey = []byte(*output.KeyMaterial)
			publicKey, err = getPublicKey(privateKey)
			if err != nil {
				return err
			}
			keyPath := m.MustStringVariable("credentials", "ssh", "key_path")
			logrus.Infof("saving private key '%v' to %v", keyName, keyPath)
			if err = ioutil.WriteFile(keyPath, privateKey, 0600); err != nil {
				return err
			}
		} else {
			logrus.Infof("importing key '%v' in region %v", keyName, region.Id)
			keyPairInput := &ec2.ImportKeyPairInput{
				KeyName:           &keyName,
				PublicKeyMaterial: publicKey,
			}
			if _, err := ec2Client.ImportKeyPair(keyPairInput); err != nil {
				return err
			}
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

	publicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(publicKey), nil
}
