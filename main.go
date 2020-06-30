package main

import (
	"flag"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const sshPort string = "22"

var passwordAuth bool = true

type path struct {
	remotePath string
	destPath   string
	mode       os.FileMode
}

func main() {
	identityFile := flag.String("i", "", "location of private key file")

	flag.Parse()
	log.SetFlags(0)

	if len(flag.Args()) < 2 {
		PrintHelp()
	}

	var err error
	remote := flag.Args()[0]
	local := flag.Args()[1]

	if remote == "" {
		log.Println("owl: remote location is mandatory parameter")
		PrintHelp()
	}

	if local == "" {
		log.Println("owl: local location is mandatory parameter")
		PrintHelp()
	}

	user, host, remoteResource := ParseRemoteString(remote)

	err, identityFilePath := GetIdentityPath(identityFile)
	if err != nil {
		log.Fatal(err)
	}

	err, config := GetSSHConfig(user, host, identityFilePath)
	if err != nil {
		log.Fatal(err)
	}

	url := host + ":" + sshPort
	conn, err := ssh.Dial("tcp", url, config)
	if err != nil {
		if strings.Contains(err.Error(), "ssh: unable to authenticate") {
			log.Fatal(`owl: Permission denied. Please enter correct password or try to authenticate with private key file by using owl -i flag`)
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

	pathSlice := walkRemotePath(remoteResource, local, client)

	for _, v := range pathSlice {
		_, err := download(v, client)
		if err != nil {
			log.Println(err)
		}
	}
}

func walkRemotePath(remoteResource, local string, client *sftp.Client) []path {
	var pathSlice []path
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
			pathSlice = append(pathSlice, path{walker.Path(), dstPath, mode})
		case mode.IsRegular():
			dstPath, _ := filepath.Split(filepath.Join(local, relPath))
			pathSlice = append(pathSlice, path{walker.Path(), dstPath, mode})
		}
	}
	return pathSlice
}

func download(v path, c *sftp.Client) (int64, error) {
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
