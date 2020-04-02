package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

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

		err := os.Link(path()+binary, path()+binaryTmp)
		if err != nil {
			panic(err)
		}
	}
}

func path() string {
	return fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())
}

func main() {
	checkEnvironment()
	prepare()
	KillOldProcess()
	defer func() {
		err := recover()
		KillOldProcess()
		if err != nil {
			panic(err)
		}
	}()

	var configPath string
	flag.StringVar(&configPath, "c", "", "ss-check -c /path/to/config.json")
	flag.Parse()
	if configPath == "" {
		fmt.Println("Usage: ss-check -c /path/to/config.json")
		os.Exit(-1)
	}

	var runner = NewRunner(configPath)
	defer runner.Report()
	defer runner.Clean()

	for _, tester := range runner.Testers() {
		go func(tester *Tester) {
			// 1. Start new ss-local process with different Config
			// 2. Test tunnel to google
			tester.Wg.Add(2)
			go tester.StartPrivoxy()
			go tester.StartSSLocal()
			tester.Wg.Wait()
			time.Sleep(time.Millisecond * 10)

			// Wait test process finished.
			tester.TestConnection(runner)
			tester.ExitSSLocal()
			tester.ExitPrivoxy()

			runner.Wg.Done()
		}(tester)
	}

	runner.Wg.Wait()
}

// 1. for 循环里面 go func(xx Type) {}(yy) go 起新的协程需要定义形参，防止后续的变量覆盖掉前面的
// 2. fmt.Sprintf("%s", 133)，这种写法错误，不会转换为 "133"
