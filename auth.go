package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/signal"
)

// https://gist.github.com/jlinoff/e8e26b4ffa38d379c7f1891fd174a6d0#file-getpassword2-go
// GetPassword provides a prompt to input a password in hidden form
func GetPassword(prompt string) (string, error) {
	// Get the initial state of the terminal.
	fd := int(os.Stdin.Fd())

	initialTermState, e1 := terminal.GetState(fd)
	if e1 != nil {
		return "", e1
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = terminal.Restore(fd, initialTermState)
		os.Exit(1)
	}()

	// Now get the password.
	fmt.Print(prompt)
	p, err := terminal.ReadPassword(fd)
	fmt.Println("")
	if err != nil {
		return "", nil
	}

	// Stop looking for ^C on the channel.
	signal.Stop(c)

	return string(p), nil
}
