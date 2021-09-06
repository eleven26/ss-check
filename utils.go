package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
)

// FileExists Check if a file exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CopyFile Copy file easily.
func CopyFile(src, dst string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(dst, input, 0755)
	if err != nil {
		panic(err)
	}
}

// HomeDir Get current user home directory.
func HomeDir() string {
	var usr, _ = user.Current()
	return usr.HomeDir
}

// KillOldProcess Kill old runner processes.
func KillOldProcess() {
	for _, binary := range []string{"ss-local-tmp", "privoxy-tmp"} {
		command := fmt.Sprintf("ps aux | grep %s | grep -v grep | awk '{print $2}' | xargs kill -9", binary)
		cmd := exec.Command("/bin/sh", "-c", command)
		_, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
	}
}
