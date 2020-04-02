package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

// Check if a file exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Copy file easily.
func CopyFile(src, dst string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(dst, input, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

// Get current user home directory.
func HomeDir() string {
	var usr, _ = user.Current()
	return usr.HomeDir
}
