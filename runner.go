package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"sync"
)

var workingDir = fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())

// Implement Sort interface.
type Runner struct {
	testers           []*Tester
	privoxyConfigPath string
	configPath        string
	serverConfigs     ServerConfigs
	Total             float64
	tested            float64
}

// Create an new runner.
func NewRunner(configPath string) *Runner {
	if !FileExists(configPath) {
		panic(fmt.Sprintf("File '%s' not exists.\n", configPath))
	}

	runner := &Runner{
		testers:    make([]*Tester, 0),
		configPath: configPath,
	}
	runner.parseConfigFile()
	runner.createTesters()

	return runner
}

func (r *Runner) Len() int {
	return len(r.testers)
}

func (r *Runner) Swap(i, j int) {
	r.testers[i], r.testers[j] = r.testers[j], r.testers[i]
}

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
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &r.serverConfigs)
	if err != nil {
		log.Fatal(err)
	}

	r.Total = float64(len(r.serverConfigs.Configs))
}

func (r *Runner) createTesters() {
	for _, config := range r.serverConfigs.Configs {
		tester := &Tester{
			Wg:     &sync.WaitGroup{},
			Config: config,
		}
		r.testers = append(r.testers, tester)
	}
}

// Generate test report.
func (r *Runner) Report() {
	// Sort test results
	sort.Sort(r)

	for _, tester := range r.testers {
		tester.Report()
		tester.Exit()
	}
}

// Start privoxy process to accept http proxy requests.
func (r *Runner) StartPrivoxy() {
	r.privoxyConfigPath = PrivoxyConfigPath()
	privoxyBinary := path() + "privoxy-tmp"
	cmd := exec.Command(privoxyBinary, r.privoxyConfigPath)
	cmd.Dir = workingDir
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("startPrivoxy Command finished with error: %v", err)
	}
}

// Kill old runner processes.
func KillOldProcess() {
	for _, binary := range []string{"ss-local-tmp", "privoxy-tmp"} {
		command := fmt.Sprintf("ps aux | grep %s | grep -v grep | awk '{print $2}' | xargs kill -9", binary)
		cmd := exec.Command("/bin/sh", "-c", command)
		_, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Remove temporary files. Stop testing processes.
func (r *Runner) Clean() {
	KillOldProcess()

	// Delete temporary privoxy config file.
	if FileExists(r.privoxyConfigPath) {
		err := os.Remove(r.privoxyConfigPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Remove %s fails: %+v\n", r.privoxyConfigPath, err))
		}
	}

	// Delete temporary ss-local-tmp, privoxy-tmp files/
	for _, binary := range []string{"ss-local-tmp", "privoxy-tmp"} {
		if FileExists(workingDir + "/" + binary) {
			err := os.Remove(binary)
			if err != nil {
				fmt.Println(fmt.Sprintf("Remove %s fails: %+v\n", workingDir + "/" + binary, err))
			}
		}
	}
}

// Get the underlying testers.
func (r *Runner) Testers() []*Tester {
	return r.testers
}
