package main

type ServerConfigs struct {
	Configs   []Config `json:"configs"`
	LocalPort int      `json:"local_port"`
}

type Config struct {
	Enable       bool   `json:"enable"`
	Password     string `json:"password"`
	Method       string `json:"method"`
	Remarks      string `json:"remarks"`
	Server       string `json:"server"`
	Obfs         string `json:"obfs"`
	Protocol     string `json:"protocol"`
	ServerPort   int    `json:"server_port"`
	RemarkBase64 string `json:"remark_base_64"`
}

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

// Convert Config to SSLocalConfig
func (s *ServerConfigs) ToSSLocalConfig(config Config) SSLocalConfig {
	return SSLocalConfig{
		ProtocolParam: "",
		Method:        config.Method,
		Protocol:      config.Protocol,
		Server:        config.Server,
		Password:      config.Password,
		LocalAddress:  "0.0.0.0",
		ServerPort:    config.ServerPort,
		Timeout:       60,
		LocalPort:     56321,
		Obfs:          config.Obfs,
		ObfsParam:     "",
	}
}
