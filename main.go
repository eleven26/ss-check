package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Tester struct {
	SSLocalProcess *os.Process
	PrivoxyProcess *os.Process
	TestProcess    *os.Process
	Wg             *sync.WaitGroup

	SSLocalPid int
	PrivoxyPid int
}

// Check if a file exists
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func testConnection(tester *Tester, wg *sync.WaitGroup) {
	testBinary := "export http_proxy=http://127.0.0.1:58321;export https_proxy=http://127.0.0.1:58321;curl -m 2 https://www.google.com"
	cmd := exec.Command("/bin/sh", "-c", testBinary)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	tester.TestProcess = cmd.Process
	wg.Done()
	fmt.Println(fmt.Sprint(cmd.Stdout))
	fmt.Println(fmt.Sprint(cmd.Stderr))
	err = cmd.Wait()
	log.Printf("testConnection Command finished with error: %v", err)
}

func startSSLocal(configPath string, tester *Tester) {
	ssLocalBinary := path() + "ss-local"
	cmd := exec.Command(ssLocalBinary, "-c", configPath)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	//tester.SSLocalProcess = cmd.Process
	tester.SSLocalPid = cmd.Process.Pid
	fmt.Println(11)
	tester.Wg.Done()
	err = cmd.Wait()
	log.Printf("startSSLocal Command finished with error: %v", err)
}

func startPrivoxy(configPath string, tester *Tester) {
	privoxyBinary := path() + "privoxy"
	cmd := exec.Command(privoxyBinary, "-c", configPath)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	//tester.PrivoxyProcess = cmd.Process
	tester.PrivoxyPid = cmd.Process.Pid
	fmt.Println(22)
	tester.Wg.Done()
	err = cmd.Wait()
	log.Printf("startPrivoxy Command finished with error: %v", err)
}

func path() string {
	return "/Users/ruby/Library/Application Support/ShadowsocksX-NG/"
}

func main() {
	serverConfigs := ServerConfigs{}
	tester := Tester{Wg: &sync.WaitGroup{}}

	// Get config file path
	configPath := flag.String("c", "", "ss-local json config file path")
	flag.Parse()
	if *configPath == "" {
		fmt.Println("Usage: ss-local -c /path/to/config.json")
		os.Exit(-1)
	} else {
		// Check if config file exists
		if !fileExists(*configPath) {
			log.Fatalf("%s not exists.\n", *configPath)
		}

		// parse config file
		content, err := ioutil.ReadFile(*configPath)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(content, &serverConfigs)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("%#v\n", serverConfigs.Configs)

	// Get temp file to save config for a ss server
	tmpFile, err := ioutil.TempFile("", "configs")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	var wg sync.WaitGroup
	// Write config to tmp file
	for _, config := range serverConfigs.Configs {
		serverConfig, err := json.Marshal(serverConfigs.ToSSLocalConfig(config))
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(tmpFile.Name(), serverConfig, 0644)
		if err != nil {
			log.Fatal(err)
		}

		// 1. Start new ss-local and privoxy process with different config
		// 2. Test tunnel to google
		//go startSSLocal(tmpFile.Name(), tester)

		// Wait process started.
		tester.Wg.Add(2)
		fmt.Println(1)
		wg.Add(1)
		go startSSLocal("/Users/ruby/Code/ss-check/server.json", &tester)
		go startPrivoxy("/Users/ruby/Code/ss-check/privoxy.config", &tester)
		tester.Wg.Wait()
		fmt.Println(2)

		// Wait test process finished.
		go testConnection(&tester, &wg)
		wg.Wait()

		//fmt.Println("SSLocalProcess:", tester.SSLocalProcess.Pid)
		//fmt.Println("PrivoxyProcess:", tester.PrivoxyProcess.Pid)
		//tester.SSLocalProcess.Kill()
		//tester.PrivoxyProcess.Kill()

		fmt.Println("SSLocalPid:", tester.SSLocalPid)
		fmt.Println("PrivoxyPid:", tester.PrivoxyPid)
		err = syscall.Kill(tester.SSLocalPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("ss-local abnormal exit")
		}
		err = syscall.Kill(tester.PrivoxyPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("privoxy abnormal exit")
		}

		fmt.Println("tmp file: ", tmpFile.Name())
	}
}
