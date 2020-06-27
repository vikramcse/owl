package main

import (
	"bufio"
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

func GetIdentityPath(identityFile *string) (error, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("owl: Not able to get current user (%v)", err), ""
	}

	identityFilePath := filepath.Join(homeDir, ".ssh", "id_rsa")

	if *identityFile != "" {
		passwordAuth = false
		identityFilePath = *identityFile
	}
	return err, identityFilePath
}

func GetSSHConfig(user, host, identityFilePath string) (error, *ssh.ClientConfig) {
	var config *ssh.ClientConfig
	var err error
	if passwordAuth {
		password, err := GetPassword(fmt.Sprintf("%s@%s's password: ", user, host))
		if err != nil {
			return err, config
		}

		if password == "" {
			return fmt.Errorf("owl: password can not be empty"), config
		}

		config, err = GetPasswordConfig(host, user, password)
	} else {
		config, err = GetPublicKeyConfig(host, user, identityFilePath)
	}

	return err, config
}
