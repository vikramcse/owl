package main

import (
	"flag"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const sshPort string = "22"

var passwordAuth bool = true

type Path struct {
	remotePath string
	destPath   string
	mode       os.FileMode
}

func main() {
	identityFlag := flag.Bool("i", false, "authenticate with private key")
	identityFile := flag.String("if", "", "location of private key file")

	flag.Parse()
	log.SetFlags(0)

	remote := flag.Args()[0]
	local := flag.Args()[1]

	if *identityFlag || *identityFile != "" {
		passwordAuth = false
		if *identityFile == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("owl: Not able to get current user (%v)", err)
			}

			*identityFile = filepath.Join(homeDir, ".ssh", "id_rsa")
		}
	}

	if remote == "" {
		log.Fatal("owl: remote location is mandatory parameter")
	}

	var err error
	user, host, remoteResource, err := ParseRemoteString(remote)
	if err != nil {
		log.Fatal(err)
	}

	var config *ssh.ClientConfig
	if passwordAuth {
		password, err := GetPassword(fmt.Sprintf("%s@%s's password: ", user, host))
		if err != nil {
			log.Fatal(err)
		}

		if password == "" {
			log.Fatal("owl: password can not be empty")
		}

		config, err = GetPasswordConfig(host, user, password)
	} else {
		config, err = GetPublicKeyConfig(host, user, *identityFile)
	}

	if err != nil {
		log.Fatal(err)
	}

	url := host + ":" + sshPort
	conn, err := ssh.Dial("tcp", url, config)
	if err != nil {
		if strings.Contains(err.Error(), "ssh: unable to authenticate") {
			log.Fatalf("Permission denied, please try again: %s", err)
		}

		log.Fatalf("owl: Failed to dial: %s", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatalf("owl: Filed to close ssh connection")
		}
	}()

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalf("owl: Filed to close sftp client connection")
		}
	}()

	log.Printf("Connecting to %s", host)
	log.Printf("Fetching %v to %v", remoteResource, local)

	var pathSlice []Path
	rootDir := filepath.Base(remoteResource)
	walker := client.Walk(remoteResource)

	for walker.Step() {
		basePath := strings.Split(walker.Path(), rootDir)[0]
		relPath, err := filepath.Rel(basePath, walker.Path())
		if err != nil {
			log.Fatal(err)
		}

		dstPath := filepath.Join(local, relPath)
		switch mode := walker.Stat().Mode(); {
		case mode.IsDir():
			pathSlice = append(pathSlice, Path{walker.Path(), dstPath, mode})
		case mode.IsRegular():
			dstPath, _ := filepath.Split(filepath.Join(local, relPath))
			pathSlice = append(pathSlice, Path{walker.Path(), dstPath, mode})
		}
	}

	downloadFileSeq(pathSlice, client)
}

func downloadFileSeq(pathSlice []Path, c *sftp.Client) {
	for _, v := range pathSlice {
		_, err := download(v, c)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func download(v Path, c *sftp.Client) (int64, error) {
	var n int64

	switch {
	case v.mode.IsDir():
		if err := os.MkdirAll(v.destPath, v.mode); err != nil && !os.IsExist(err) {
			return 0, err
		}
		log.Printf("Retriving %s", v.remotePath)
	case v.mode.IsRegular():
		rSrcFile, err := c.Open(v.remotePath)
		if err != nil {
			return 0, err
		}

		rSrcFileInfo, _ := rSrcFile.Stat()

		n, err = fileCopy(rSrcFile, v.destPath, rSrcFileInfo)
		if err != nil {
			log.Printf("%v \n", err)
			return 0, err
		}

		log.Printf("%s", v.remotePath)
	}

	return n, nil
}
