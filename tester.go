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

	tmpFile           *os.File
	privoxyConfigPath string
	httpPort          int
	socksPort         int
}

func NewTester(httpPort, socksPort int) *Tester {
	return &Tester{
		Wg:        &sync.WaitGroup{},
		httpPort:  httpPort,
		socksPort: socksPort,
	}
}

func (t *Tester) Usable() string {
	if t.IsUsable {
		return " ✔ "
	} else {
		return " ✘ "
	}
}

func (t *Tester) Exit() {
	t.ExitSSLocal()
	t.ExitPrivoxy()

	if t.tmpFile != nil && FileExists(t.tmpFile.Name()) {
		err := os.Remove(t.tmpFile.Name())
		if err != nil {
			fmt.Printf("Remove %s fails: %+v\n", t.tmpFile.Name(), err)
		}
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

func (t *Tester) ExitPrivoxy() {
	if t.PrivoxyPid > 0 {
		err := syscall.Kill(t.PrivoxyPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("privoxy abnormal exit")
		}
		t.PrivoxyPid = 0
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

	serverConfig, err := json.Marshal(ToSSLocalConfig(t.Config, t.socksPort))
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(tmpFile.Name(), serverConfig, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name()
}

func (t *Tester) TestConnection(runner *Runner) {
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-m", "1", "http://www.google.com")
	//cmd := exec.Command("curl", "-m", "2", "http://www.google.com")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", t.httpPort))
	startAt := time.Now()
	out, err := cmd.Output()
	endAt := time.Now()
	t.Delay = endAt.Sub(startAt).Milliseconds()
	runner.Tested = runner.Tested + 1
	if err != nil {
		//log.Printf("TestConnection output with error: %v", err)
	}
	// Privoxy proxy error
	t.IsUsable = string(out) == "200"
	fmt.Println(t.Usable(), t.server(), t.elapsed(), fmt.Sprintf("%.2f%% (%d/%d)", runner.Tested*100.0/runner.Total, int64(runner.Tested), int64(runner.Total)))
	if err != nil {
		//log.Printf("TestConnection Command finished with error: %v", err)
	}
}

func (t *Tester) StartSSLocal() {
	ssLocalBinary := path() + "ss-local-tmp"
	cmd := exec.Command(ssLocalBinary, "-c", t.configPath())
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s/Library/Application Support/ShadowsocksX-NG/", HomeDir()))
	cmd.Dir = fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.SSLocalPid = cmd.Process.Pid
	t.Wg.Done()
	err = cmd.Wait()
	if err != nil {
		//log.Printf("StartSSLocal Command finished with error: %v\n", err)
	}
}

// Start privoxy process to accept http proxy requests.
func (t *Tester) StartPrivoxy() {
	t.privoxyConfigPath = PrivoxyConfigPath(t.httpPort, t.socksPort)
	privoxyBinary := path() + "privoxy-tmp"
	cmd := exec.Command(privoxyBinary, t.privoxyConfigPath)
	cmd.Dir = workingDir
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.Wg.Done()
	err = cmd.Wait()
	if err != nil {
		//fmt.Printf("startPrivoxy Command finished with error: %v", err)
	}
}

func (t *Tester) Report() {
	fmt.Printf("%s %s, delay: %dms\n", strings.Trim(t.Usable(), " "), t.server(), t.Delay)
}
