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

var serverConfigs = ServerConfigs{}

type Tester struct {
	Wg *sync.WaitGroup

	SSLocalPid int
	PrivoxyPid int

	IsUsable bool
	config   Config
	tmpFile  *os.File
}

func (t *Tester) exit() {
	if t.SSLocalPid > 0 {
		err := syscall.Kill(t.SSLocalPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("ss-local abnormal exit")
		}
	}

	if t.PrivoxyPid > 0 {
		err := syscall.Kill(t.PrivoxyPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("privoxy abnormal exit")
		}
	}

	if t.tmpFile != nil {
		os.Remove(t.tmpFile.Name())
	}
}

func (t *Tester) server() string {
	return t.config.Server
	//decoded, err := base64.StdEncoding.DecodeString(t.config.RemarkBase64)
	//if err != nil {
	//	return ""
	//}
	//return string(decoded)
}

func (t *Tester) configPath() string {
	// Get temp file to save config for a ss server
	tmpFile, err := ioutil.TempFile("", "configs")
	if err != nil {
		log.Fatal(err)
	}
	t.tmpFile = tmpFile

	serverConfig, err := json.Marshal(serverConfigs.ToSSLocalConfig(t.config))
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(tmpFile.Name(), serverConfig, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(serverConfig))

	return tmpFile.Name()
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

func (t *Tester) testConnection(wg *sync.WaitGroup) {
	cmd := exec.Command("curl", "-m", "2", "http://www.google.com")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "http_proxy=http://127.0.0.1:58321")
	out, err := cmd.Output()
	if err != nil {
		log.Printf("testConnection output with error: %v", err)
	}
	t.IsUsable = true
	fmt.Println(string(out))
	wg.Done()
	log.Printf("testConnection Command finished with error: %v", err)
}

func (t *Tester) startSSLocal() {
	ssLocalBinary := path() + "ss-local"
	cmd := exec.Command(ssLocalBinary, "-c", t.configPath())
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DYLD_LIBRARY_PATH=/Users/ruby/Library/Application Support/ShadowsocksX-NG/")
	cmd.Dir = "/Users/ruby/Library/Application Support/ShadowsocksX-NG/"
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.SSLocalPid = cmd.Process.Pid
	t.Wg.Done()
	//err = cmd.Wait()
	//log.Printf("startSSLocal Command finished with error: %v", err)
}

func (t *Tester) startPrivoxy(configPath string) {
	privoxyBinary := path() + "privoxy"
	cmd := exec.Command(privoxyBinary, configPath)
	cmd.Dir = "/Users/ruby/Library/Application Support/ShadowsocksX-NG/"
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.PrivoxyPid = cmd.Process.Pid
	t.Wg.Done()
	//err = cmd.Wait()
	//log.Printf("startPrivoxy Command finished with error: %v", err)
}

func killOldProcess() {
	cmd := exec.Command("/bin/sh", "-c", "ps aux | grep ss-local | grep -v grep | awk '{print $2}' | xargs kill -9")
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	cmd = exec.Command("/bin/sh", "-c", "ps aux | grep privoxy | grep -v grep | awk '{print $2}' | xargs kill -9")
	out, err = cmd.Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func path() string {
	return "/Users/ruby/Library/Application Support/ShadowsocksX-NG/"
}

func main() {
	killOldProcess()
	defer killOldProcess()

	// Get config file path
	configPath := flag.String("c", "", "ss-local json config file path")
	flag.Parse()
	if *configPath == "" {
		fmt.Println("Usage: ss-check -c /path/to/config.json")
		os.Exit(-1)
	} else {
		// Check if config file exists
		if !fileExists(*configPath) {
			log.Fatalf("%s not exists.\n", *configPath)
		}

		// Read config file
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

	var testers []Tester
	defer func() {
		for _, tester := range testers {
			fmt.Printf("server: %s, usable: %+v\n", tester.server(), tester.IsUsable)
			tester.exit()
		}
	}()

	// Write config to tmp file
	for _, config := range serverConfigs.Configs {
		var wg sync.WaitGroup
		wg.Add(1)

		tester := Tester{
			Wg:     &sync.WaitGroup{},
			config: config,
		}
		testers = append(testers, tester)

		// 1. Start new ss-local and privoxy process with different config
		// 2. Test tunnel to google

		// Wait process started.
		tester.Wg.Add(2)
		go tester.startSSLocal()
		go tester.startPrivoxy("/Users/ruby/Code/ss-check/privoxy.config")
		tester.Wg.Wait()

		// Wait test process finished.
		go tester.testConnection(&wg)
		wg.Wait()
	}
}
