package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
)

// Privoxy and ss-local working directory.
var workingDir = fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())

// We need to start multiple process to implement parallel test.
// So we need different port for privoxy and ss-local.
var httpPort = 58321
var socksPort = 56321

// Implement Sort interface.
type Runner struct {
	Total  float64
	Tested float64
	Wg     *sync.WaitGroup

	testers           []*Tester
	privoxyConfigPath string
	configPath        string
	serverConfigs     ServerConfigs
}

// Create an new runner.
func NewRunner(configPath string) *Runner {
	if !FileExists(configPath) {
		panic(fmt.Sprintf("File '%s' not exists.\n", configPath))
	}

	runner := &Runner{
		Wg:         &sync.WaitGroup{},
		configPath: configPath,
	}
	runner.parseConfigFile()
	runner.createTesters()

	runner.Wg.Add(runner.Len())

	return runner
}

// Tester count
func (r *Runner) Len() int {
	return len(r.testers)
}

// Swap tester
func (r *Runner) Swap(i, j int) {
	r.testers[i], r.testers[j] = r.testers[j], r.testers[i]
}

// Compare tester for ordering.
func (r *Runner) Less(i, j int) bool {
	var b2i = map[bool]int8{false: 0, true: 1}

	// Sort by Delay, put minimum in the end.
	if r.testers[i].IsUsable == r.testers[j].IsUsable {
		return r.testers[i].Delay > r.testers[j].Delay
	}

	// Sort by Usable, put Usable Config in the end.
	return b2i[r.testers[i].IsUsable] < b2i[r.testers[j].IsUsable]
}

// See privoxy.config
func (r *Runner) parseConfigFile() {
	// Read Config file
	content, err := ioutil.ReadFile(r.configPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &r.serverConfigs)
	if err != nil {
		panic(err)
	}

	r.Total = float64(len(r.serverConfigs.Configs))
}

// Create testers by shadowsocks configs.
func (r *Runner) createTesters() {
	r.testers = make([]*Tester, r.Len())
	for _, config := range r.serverConfigs.Configs {
		tester := NewTester(httpPort, socksPort)
		tester.Config = config
		r.testers = append(r.testers, tester)

		httpPort = httpPort + 1
		socksPort = socksPort + 1
	}
}

// Generate test report.
func (r *Runner) Report() {
	// Sort test results
	sort.Sort(r)

	for _, tester := range r.testers {
		tester.Report()
	}
}

// Remove temporary binary files.
func (r *Runner) removeTmpBinaries()  {
	// Delete temporary ss-local-tmp, privoxy-tmp files/
	for _, binary := range []string{"ss-local-tmp", "privoxy-tmp"} {
		if FileExists(workingDir + binary) {
			err := os.Remove(workingDir + binary)
			if err != nil {
				fmt.Println(fmt.Sprintf("Remove %s fails: %+v", workingDir+"/"+binary, err))
			}
		}
	}
}

// Remove temporary files. Stop testing processes.
func (r *Runner) Clean() {
	r.removeTmpBinaries()

	// Delete temporary files.
	for _, tester := range r.testers {
		tester.Clean()
		tester.Exit()
	}
}

// Get the underlying testers.
func (r *Runner) Testers() []*Tester {
	return r.testers
}
