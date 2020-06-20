package main

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const sshPort string = "22"
const passwordAuth bool = true

func main() {
	remote := os.Args[1]
	local := os.Args[2]

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
		password, err := GetPassword("password: ")
		if err != nil {
			log.Fatal(err)
		}

		if password == "" {
			log.Fatal("owl: password can not be empty")
		}

		config, err = GetPasswordConfig(host, user, password)
	} else {
		config, err = GetPublicKeyConfig(host, user)
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	url := host + ":" + sshPort
	conn, err := ssh.Dial("tcp", url, config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
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

	// open remote file
	srcFile, err := client.Open(remoteResource)
	if err != nil {
		log.Fatal(err)
	}

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		log.Fatalf("owl: Not able to get stat of remote file (%v)", err)
	}

	pathMap := make(map[string]bool)

	if !srcFileInfo.IsDir() {
		err := fileCopy(srcFile, local, srcFileInfo)
		if err != nil {
			log.Fatal(err)
		}
	} else {
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
				//if err2 := os.MkdirAll(dstPath, mode); err2 != nil && !os.IsExist(err) {
				//	log.Fatal(err2)
				//}
				pathMap[dstPath] = true
			case mode.IsRegular():
				//rSrcFile, err2 := client.Open(walker.Path())
				//if err2 != nil {
				//	log.Fatal(err2)
				//}
				//
				//rSrcFileInfo, _ := rSrcFile.Stat()
				//dstPath, _ := filepath.Split(filepath.Join(local, relPath))
				//err := fileCopy(rSrcFile, dstPath, rSrcFileInfo)
				//if err != nil {
				//	log.Fatal(err)
				//}
				pathMap[filepath.Join(local, relPath)] = false
			}
		}
	}

	fmt.Println(pathMap)
}
