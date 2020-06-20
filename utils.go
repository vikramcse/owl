package main

import "strings"

// scp username@destination_host:source destination

// ParseRemoteString will parse the source string format to get
// extract username, host and file/folder names
func ParseRemoteString(s string) (username string, host string, resource string, err error) {
	ss := strings.Split(s, "@")
	hrs := strings.Split(ss[1], ":")
	username = ss[0]
	host = hrs[0]
	resource = hrs[1]
	return
}
