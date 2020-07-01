# Owl

Owl is a command line program to download files from a remote server.

## Why
There are well build tools such as sftp, scp which provide same functionality as Owl, 
but this is how you learn new a programming language. 

## Getting Started
##### To build the binary file on your local machine
go get github.com/vikramcse/owl

##### Download pre-built binary files
1. [windows](https://github.com/vikramcse/owl/blob/master/bin/owl.exe)
2. [Linux](https://github.com/vikramcse/owl/blob/master/bin/owl)
3. On linux machine you have to make the binary executable `chmod +x owl`

##### How to download file from a remote server into your local machine
```
owl:
A SFTP like clone written in go to download files from remote server Usage:
  owl [option] user@remote:/path/to/the/file destination
Options:
  -i       path to private key file
```