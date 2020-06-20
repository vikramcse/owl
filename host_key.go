package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io/ioutil"
	"os"
	"path/filepath"
)

// getHostkeyCallback reads the known_hosts file in .ssh directory
func getHostKeyCallback(host string) (ssh.HostKeyCallback, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("owl: Not able to get current user (%v)", err)
	}

	knownHostPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	hostKeyCallback, err := knownhosts.New(knownHostPath)
	if err != nil {
		return nil, fmt.Errorf("owl: Not able to get the known hosts callback function (%v)", err)
	}

	return hostKeyCallback, nil
}

func getSignerFromPrivateKey() (ssh.Signer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("owl: Not able to get current user (%v)", err)
	}

	sshDirectory := filepath.Join(homeDir, ".ssh", "id_rsa")

	privateKey, err := ioutil.ReadFile(sshDirectory)
	if err != nil {
		return nil, fmt.Errorf("owl: Not able to read private key (%v)", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("own: Not able to parse private key (%v)", err)
	}

	return signer, nil
}

func GetPublicKeyConfig(host, user string) (*ssh.ClientConfig, error) {
	hostKeyCallback, err := getHostKeyCallback(host)
	if err != nil {
		return nil, err
	}

	signer, err := getSignerFromPrivateKey()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}

	return config, nil
}

func GetPasswordConfig(host, user, password string) (*ssh.ClientConfig, error) {
	hostKeyCallback, err := getHostKeyCallback(host)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: hostKeyCallback,
	}

	return config, nil
}

func generateRSAKeys(writeInFile bool) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	var publicKey *rsa.PublicKey

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	publicKey = &privateKey.PublicKey

	if writeInFile {
		var RSAPrivateFile string = "key.pem"

		err = func(privateKey *rsa.PrivateKey, keyfile string) error {
			file, err := os.OpenFile(keyfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				return err
			}
			defer file.Close()

			privateBlock := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}

			pem.Encode(file, privateBlock)

			return nil
		}(privateKey, RSAPrivateFile)
	}

	return publicKey, privateKey, err
}
