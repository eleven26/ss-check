package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

// Parse Config file export from ShadowsocksX-NG
type ServerConfigs struct {
	Configs   []Config `json:"configs"`
	LocalPort int      `json:"local_port"`
}

// Config parsed from exported file.
type Config struct {
	Enable        bool   `json:"enable"`
	Password      string `json:"password"`
	Method        string `json:"method"`
	Remarks       string `json:"remarks"`
	Server        string `json:"server"`
	Obfs          string `json:"obfs"`
	Protocol      string `json:"protocol"`
	ServerPort    int    `json:"server_port"`
	RemarkBase64  string `json:"remark_base_64"`
	ProtocolParam string `json:"protocolparam"`
	ObfsParam     string `json:"obfsparam"`
}

// Config file content for ss-local command.
type SSLocalConfig struct {
	ProtocolParam string `json:"protocol_param"`
	Method        string `json:"method"`
	Protocol      string `json:"protocol"`
	Server        string `json:"server"`
	Password      string `json:"password"`
	LocalAddress  string `json:"local_address"`
	ServerPort    int    `json:"server_port"`
	Timeout       int    `json:"timeout"`
	LocalPort     int    `json:"local_port"`
	Obfs          string `json:"obfs"`
	ObfsParam     string `json:"obfs_param"`
}

// Convert Config to SSLocalConfig.
func ToSSLocalConfig(config Config) SSLocalConfig {
	return SSLocalConfig{
		ProtocolParam: config.ProtocolParam,
		Method:        config.Method,
		Protocol:      config.Protocol,
		Server:        config.Server,
		Password:      config.Password,
		LocalAddress:  "0.0.0.0",
		ServerPort:    config.ServerPort,
		Timeout:       60,
		LocalPort:     56321,
		Obfs:          config.Obfs,
		ObfsParam:     config.ObfsParam,
	}
}

// Use temporary file to save privoxy Config file.
func PrivoxyConfigPath() string {
	httpProxyPort := "58321"
	socksListenPort := "56321"

	config := `listen-address 0.0.0.0:%s
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

forward-socks5 / 0.0.0.0:%s .
forward         192.168.*.*/     .
forward         10.*.*.*/        .
forward         127.*.*.*/       .`
	config = fmt.Sprintf(config, httpProxyPort, socksListenPort)

	tmpFile, err := ioutil.TempFile("", "privoxy.Config")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(tmpFile.Name(), []byte(config), 0644)
	if err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name()
}
