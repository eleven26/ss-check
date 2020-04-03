package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

var binaries = []string{"ss-local", "privoxy"}

// Check if environment support.
func checkEnvironment() {
	path := path()
	if !FileExists(path) {
		panic(fmt.Sprintf("%s not exists. Exiting...", path))
	}

	for _, binary := range binaries {
		if !FileExists(path + binary) {
			panic(fmt.Sprintf("%s not exists. Exiting...", path+binary))
		}
	}
}

// Copy binary file, so we can kill these test processes by a new name.
func copyBinaries() {
	for _, binary := range binaries {
		binaryTmp := binary + "-tmp"
		if FileExists(path() + binaryTmp) {
			err := os.Remove(path() + binaryTmp)
			if err != nil {
				panic(err)
			}
		}

		err := os.Link(path()+binary, path()+binaryTmp)
		if err != nil {
			panic(err)
		}
	}
}

func prepare() {
	checkEnvironment()
	copyBinaries()
	KillOldProcess()
}

func clean() {
	KillOldProcess()
}

func path() string {
	return fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func main() {
	prepare()
	defer clean()

	var configPath string
	var u string
	flag.StringVar(&configPath, "c", "", "ss-check -c /path/to/config.json")
	flag.StringVar(&u, "u", "", "ss-check -u www.google.com(only domain)")
	flag.Parse()
	if configPath == "" {
		fmt.Println("Usage: ss-check -c /path/to/config.json <-u www.google.com(only domain)>")
		os.Exit(-1)
	}
	if u != "" {
		if !IsUrl(u) {
			panic(fmt.Sprintf("%s is not a valid URL.\n", u))
		}
	} else {
		u = "www.google.com"
	}

	var runner = NewRunner(configPath)
	defer runner.Report()
	defer runner.Clean()

	// Maximum running processes num is 5 * 2 (privoxy + ss-local)
	throttle := make(chan int, 5)

	for _, tester := range runner.Testers() {
		go func(tester *Tester) {
			throttle<-1
			defer func() {
				tester.Exit()
				runner.Wg.Done()
				<-throttle
			}()

			// 1. Start new privoxy and ss-local process with different config
			// 2. Test http tunnel to google
			tester.Wg.Add(2)
			go tester.StartPrivoxy()
			go tester.StartSSLocal()
			tester.Wg.Wait()

			tester.TestConnection(runner, u)
		}(tester)
	}

	runner.Wg.Wait()
}

// 1. for 循环里面 go func(xx Type) {}(yy) go 起新的协程需要定义形参，防止后续的变量覆盖掉前面的
// 2. fmt.Sprintf("%s", 133)，这种写法错误，不会转换为 "133"
