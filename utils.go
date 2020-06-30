package main

import (
	"fmt"
	"os"
	"strings"
)

// scp username@destination_host:source

// ParseRemoteString will parse the source string format to get
// extract username, host and file/folder names
func ParseRemoteString(s string) (username string, host string, resource string) {
	ss := strings.Split(s, "@")
	hrs := strings.Split(ss[1], ":")
	username = ss[0]
	host = hrs[0]
	resource = hrs[1]
	return
}

func PrintHelp() {
	fmt.Print(`owl: 
A SFTP like clone written in go to download files from remote server Usage:
  owl [option] user@remote:/path/to/the/file destination
Options:
  -i       path to private key file
`)
	os.Exit(1)
}
