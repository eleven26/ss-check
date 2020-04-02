package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Tester struct {
	Wg         *sync.WaitGroup
	SSLocalPid int
	PrivoxyPid int
	IsUsable   bool
	Config     Config
	Delay      int64

	tmpFile *os.File
}

func (t *Tester) Usable() string {
	if t.IsUsable {
		return " ✔ "
	} else {
		return " ✘ "
	}
}

func (t *Tester) Exit() {
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

func (t *Tester) ExitSSLocal() {
	if t.SSLocalPid > 0 {
		err := syscall.Kill(t.SSLocalPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("ss-local abnormal exit")
		}
		t.SSLocalPid = 0
	}
}

func (t *Tester) server() string {
	return strings.Trim(t.Config.Server, " ") + fmt.Sprintf("(%s)", t.Config.Remarks)
}

func (t *Tester) elapsed() string {
	return fmt.Sprintf("(%dms)", t.Delay)
}

func (t *Tester) configPath() string {
	// Get temp file to save Config for a ss server
	tmpFile, err := ioutil.TempFile("", "configs")
	if err != nil {
		log.Fatal(err)
	}
	t.tmpFile = tmpFile

	serverConfig, err := json.Marshal(ToSSLocalConfig(t.Config))
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(tmpFile.Name(), serverConfig, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name()
}

func (t *Tester) TestConnection(ch chan<- bool, runner *Runner) {
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-m", "1", "http://www.google.com")
	//cmd := exec.Command("curl", "-m", "2", "http://www.google.com")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "http_proxy=http://127.0.0.1:58321")
	startAt := time.Now()
	out, err := cmd.Output()
	endAt := time.Now()
	t.Delay = endAt.Sub(startAt).Milliseconds()
	tested = tested + 1
	if err != nil {
		//log.Printf("TestConnection output with error: %v", err)
	}
	// Privoxy proxy error
	t.IsUsable = string(out) == "200"
	fmt.Println(t.Usable(), t.server(), t.elapsed(), fmt.Sprintf("%.2f%% (%d/%d)", tested*100.0/runner.Total, int64(tested), int64(runner.Total)))
	ch <- true
	if err != nil {
		//log.Printf("TestConnection Command finished with error: %v", err)
	}
}

func (t *Tester) StartSSLocal() {
	ssLocalBinary := path() + "ss-local-tmp"
	cmd := exec.Command(ssLocalBinary, "-c", t.configPath())
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s/Library/Application Support/ShadowsocksX-NG/", home))
	cmd.Dir = fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", home)
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.SSLocalPid = cmd.Process.Pid
	t.Wg.Done()
	err = cmd.Wait()
	if err != nil {
		//log.Printf("StartSSLocal Command finished with error: %v", err)
	}
}

func (t *Tester) Report() {
	fmt.Printf("%s %s, Delay: %dms\n", strings.Trim(t.Usable(), " "), t.server(), t.Delay)
}
