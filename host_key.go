package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getHostKeyCallback(host string) (ssh.HostKeyCallback, error) {
	var hostKeyCallback ssh.HostKeyCallback
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return hostKeyCallback, fmt.Errorf("owl: Not able to get current user (%v)", err)
	}

	knownHostPath := filepath.Join(homeDir, ".ssh", "known_hosts")
	file, err := os.Open(knownHostPath)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}

		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return hostKeyCallback, fmt.Errorf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		return hostKeyCallback, fmt.Errorf("owl: No hostkey present for host %s. To generate host key please generate host key using ssh-keygen or ssh user@hostname", host)
	}

	if err := file.Close(); err != nil {
		return hostKeyCallback, fmt.Errorf("owl: Not able to close file (%v)", err)
	}

	return ssh.FixedHostKey(hostKey), nil
}

func getSignerFromPrivateKey(identityFile string) (ssh.Signer, error) {
	privateKey, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("owl: Not able to read private key (%v)", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("own: Not able to parse private key (%v)", err)
	}

	return signer, nil
}

func GetPublicKeyConfig(host, user, identityFile string) (*ssh.ClientConfig, error) {
	hostKeyCallback, err := getHostKeyCallback(host)
	if err != nil {
		return nil, err
	}

	signer, err := getSignerFromPrivateKey(identityFile)
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
