package common

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config stores the data required for the nodes to run.
type Config struct {
	Address AddressConfig `json:"address"`
	Irc     IrcConfig     `json:"irc"`
	Master  MasterConfig  `json:"master"`
	Node    ConnInfo      `json:"node"`
}

type ConnInfo struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
}

func (conn *ConnInfo) String() string {
	return conn.Hostname + ":" + strconv.Itoa(conn.Port)
}

// AddressConfig stores the addresses of where the load balancers will be
//   running at.
type AddressConfig struct {
	App    ConnInfo `json:"app"`
	Reader ConnInfo `json:"reader"`
	Writer ConnInfo `json:"writer"`
}

// ConnConfig stores the data required for a node to accept work.
type MasterConfig struct {
	NodeRegistryPath   string `json:"nodeRegistryPath"`
	NodeCheckFrequency int    `json:"nodeCheckFrequency"`
}

// IrcConfig stores the data required for a node to use IRC.
type IrcConfig struct {
	MessageLimit  int    `json:"messageLimit"`
	ReadFrequency int    `json:"readFrequency"`
	Nickname      string `json:"nickname"`
	Password      string `json:"password"`
	ConnInfo      string `json:"connectionInfo"`
}

// LoadConfig returns the configuration read into a Config struct.
func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	parser := json.NewDecoder(file)
	err = parser.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
