package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Tester Attributes of Tester:
// 		Wg: WaitGroup for manipulate privoxy and ss-local process.
// 		IsUsage: If Config is Usable
// 		Config:  Shadowsocks server config
//		Delay:   Delay of accessing http://www.google.com
//		SSLocalPid: Record process id of ss-local.
// 		PrivoxyPid: Record process id of privoxy.
//		ssLocalConfigPath: Temporary configuration file for ss-local of current config.
//    	privoxyConfigPath: Temporary configuration file for privoxy of current config.
// 		httpPort:  Local http proxy port for privoxy
//      socksPort: Port for ss-local listening.
type Tester struct {
	Wg        *sync.WaitGroup
	IsUsable  bool
	IsTimeout bool
	Config    Config
	Delay     int64

	SSLocalPid        int
	PrivoxyPid        int
	ssLocalConfigPath string
	privoxyConfigPath string
	httpPort          int
	socksPort         int
}

// NewTester Create a new tester for testing config.
func NewTester(httpPort, socksPort int) *Tester {
	return &Tester{
		Wg:        &sync.WaitGroup{},
		httpPort:  httpPort,
		socksPort: socksPort,
	}
}

// More intuitive info about usability of shadowsocks config.
func (t *Tester) usable() string {
	if t.IsUsable {
		return " ✔ "
	} else {
		return " ✘ "
	}
}

// Exit Exit privoxy and ss-local processes.
func (t *Tester) Exit() {
	t.exitSSLocal()
	t.exitPrivoxy()
}

// Exit the ss-local process of current tester.
func (t *Tester) exitSSLocal() {
	if t.SSLocalPid > 0 {
		err := syscall.Kill(t.SSLocalPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("ss-local abnormal exit")
		}
		t.SSLocalPid = 0
	}
}

// Exit the privoxy process of current tester.
func (t *Tester) exitPrivoxy() {
	if t.PrivoxyPid > 0 {
		err := syscall.Kill(t.PrivoxyPid, syscall.SIGTERM)
		if err != nil {
			fmt.Println("privoxy abnormal exit")
		}
		t.PrivoxyPid = 0
	}
}

// Delete temporary ss-local config file.
func (t *Tester) removeSSLocalConfigFile() {
	if FileExists(t.ssLocalConfigPath) {
		err := os.Remove(t.ssLocalConfigPath)
		if err != nil {
			fmt.Printf("Remove %s fails: %+v\n", t.ssLocalConfigPath, err)
		}
	}
}

// Delete temporary privoxy config file.
func (t *Tester) removePrivoxyConfigFile() {
	if FileExists(t.privoxyConfigPath) {
		err := os.Remove(t.privoxyConfigPath)
		if err != nil {
			fmt.Println(fmt.Sprintf("Remove %s fails: %+v\n", t.privoxyConfigPath, err))
		}
	}
}

// Clean temporary files.
func (t *Tester) Clean() {
	t.removeSSLocalConfigFile()
	t.removePrivoxyConfigFile()
}

// Server info, including server name/ip and server remark.
func (t *Tester) server() string {
	return strings.Trim(t.Config.Server, " ") + fmt.Sprintf("(%s)", t.Config.Remarks)
}

// The time from send request to receive response.
func (t *Tester) elapsed() string {
	return fmt.Sprintf("(%dms)", t.Delay)
}

// Create temporary file to save ss-local config.
func (t *Tester) tmpSSLocalConfigPath() string {
	// Get temp file to save Config for a ss server
	tmpFile, err := ioutil.TempFile("", "configs")
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()

	serverConfig, err := json.Marshal(ToSSLocalConfig(t.Config, t.socksPort))
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(tmpFile.Name(), serverConfig, 0644)
	if err != nil {
		panic(err)
	}

	t.ssLocalConfigPath = tmpFile.Name()
	return tmpFile.Name()
}

// Create temporary file to save privoxy config.
func (t *Tester) tmpPrivoxyConfigPath() string {
	config := `listen-address 0.0.0.0:%d
toggle  1
enable-remote-toggle 1
enable-remote-http-toggle 1
enable-edit-actions 0
enforce-blocks 0
buffer-limit 4096
forwarded-connect-retries  0
accept-intercepted-requests 0
allow-cgi-request-crunching 0
split-large-forms 0
keep-alive-timeout 5
socket-timeout 60

forward-socks5 / 0.0.0.0:%d .
forward         192.168.*.*/     .
forward         10.*.*.*/        .
forward         127.*.*.*/       .`
	config = fmt.Sprintf(config, t.httpPort, t.socksPort)

	tmpFile, err := ioutil.TempFile("", "privoxy.Config")
	if err != nil {
		panic(err)
	}
	defer tmpFile.Close()

	err = ioutil.WriteFile(tmpFile.Name(), []byte(config), 0644)
	if err != nil {
		panic(err)
	}

	t.privoxyConfigPath = tmpFile.Name()
	return tmpFile.Name()
}

// TestConnection Check if connection can access google.com.
func (t *Tester) TestConnection(runner *Runner, u string) {
	// Wait for privoxy to be ready
	time.Sleep(time.Second * 1)

	// -o <path>:redirect output
	// -w %{http_code}: output the http status code
	// -m <seconds>: timeout
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-m", "10", fmt.Sprintf("http://%s", u))
	//cmd := exec.Command("curl", "-m", "2", "http://www.google.com")

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("http_proxy=http://127.0.0.1:%d", t.httpPort))

	startAt := time.Now()
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to execute curl command: %s, status code: %v\n", err.Error(), string(out))
	}
	endAt := time.Now()
	t.Delay = endAt.Sub(startAt).Milliseconds()

	runner.Tested = runner.Tested + 1
	// 500 is basically a privoxy proxy error.
	t.IsUsable = string(out) != "500" && string(out) != "000"
	t.IsTimeout = err != nil && strings.Contains(err.Error(), "28") && string(out) == "000"

	fmt.Println(t.usable(), t.server(), t.elapsed(), fmt.Sprintf("%.2f%% (%d/%d)", runner.Tested*100.0/runner.Total, int64(runner.Tested), int64(runner.Total)))
}

// StartSSLocal Start ss-local process for connecting to shadowsocks server.
func (t *Tester) StartSSLocal() {
	ssLocalBinary := path() + "ss-local-tmp"
	cmd := exec.Command(ssLocalBinary, "-c", t.tmpSSLocalConfigPath())
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("DYLD_LIBRARY_PATH=%s/Library/Application Support/ShadowsocksX-NG/", HomeDir()))
	cmd.Dir = fmt.Sprintf("%s/Library/Application Support/ShadowsocksX-NG/", HomeDir())
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.SSLocalPid = cmd.Process.Pid
	t.Wg.Done()
	_ = cmd.Wait()
}

// StartPrivoxy Start privoxy process to accept http proxy requests.
func (t *Tester) StartPrivoxy() {
	privoxyBinary := path() + "privoxy-tmp"
	cmd := exec.Command(privoxyBinary, t.tmpPrivoxyConfigPath())
	cmd.Dir = workingDir
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	t.Wg.Done()
	_ = cmd.Wait()
}

// Report Example: ✔ example.com(remark), delay: 355ms
func (t *Tester) Report() {
	fmt.Printf("%s %s, delay: %dms", strings.Trim(t.usable(), " "), t.server(), t.Delay)
	if t.IsTimeout {
		fmt.Printf(" (timeout)")
	}
	fmt.Println()
}
