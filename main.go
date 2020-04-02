package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var home = HomeDir()
var tested = 0.0

func checkEnvironment() {
	path := path()
	if !FileExists(path) {
		panic(fmt.Sprintf("%s not exists. Exiting...", path))
	}

	for _, binary := range []string{"ss-local", "privoxy"} {
		if !FileExists(path + binary) {
			panic(fmt.Sprintf("%s not exists. Exiting...", path+binary))
		}
	}
}

func prepare() {
	for _, binary := range []string{"ss-local", "privoxy"} {
		binaryTmp := binary + "-tmp"
		if FileExists(path() + binaryTmp) {
			err := os.Remove(path() + binaryTmp)
			if err != nil {
				log.Fatal(err)
			}
		}

		os.Link(path()+binary, path()+binaryTmp)
	}
}

func path() string {
	return fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", home)
}

func main() {
	checkEnvironment()
	prepare()

	var flags struct {
		ConfigPath string
		Kill       bool
	}

	// Get Config file path
	flag.StringVar(&flags.ConfigPath, "c", "", "ss-check json Config file path")
	flag.BoolVar(&flags.Kill, "k", false, "ss-check -k")
	flag.Parse()
	if flags.ConfigPath == "" {
		fmt.Println("Usage: ss-check -c /path/to/Config.json")
		os.Exit(-1)
	}

	var runner = NewRunner(flags.ConfigPath)
	defer runner.Report()
	defer runner.Clean()

	if flags.Kill {
		KillOldProcess()
	}

	go runner.StartPrivoxy()
	time.Sleep(time.Second * 1) // Waiting for privoxy started

	// Write Config to tmp file
	for _, tester := range runner.Testers() {
		// 1. Start new ss-local process with different Config
		// 2. Test tunnel to google
		tester.Wg.Add(1)
		go tester.StartSSLocal()
		tester.Wg.Wait()
		time.Sleep(time.Millisecond * 10)

		// Wait test process finished.
		ch := make(chan bool)
		tester.TestConnection(ch, runner)
		<-ch

		tester.ExitSSLocal()
		time.Sleep(time.Millisecond * 10)
	}
	time.Sleep(time.Second)
}
