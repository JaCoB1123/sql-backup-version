package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type server struct {
	Host                     string
	Instance                 string
	User                     string
	Password                 string
	IntegratedAuthentication bool
	VersionDescription       string `json:"-"`
	Version                  string `json:"-"`
	Level                    string `json:"-"`
	Edition                  string `json:"-"`
}

type serverList []server

// Configuration saves the configuration
type configuration struct {
	Servers serverList
}

// GetConfig returns the saved configuration
func getConfig() (*configuration, error) {
	config, err := ioutil.ReadFile("./config/servers.json")
	if err != nil {
		return nil, err
	}

	var configuration configuration
	var servers serverList
	err = json.Unmarshal(config, &servers)
	if err != nil {
		return nil, err
	}

	configuration.Servers = servers

	return &configuration, nil
}

func (s server) String() string {
	return fmt.Sprintf("%s\\%s\nVersion: %s (%s)\nEdition: %s", s.Host, s.Instance, s.Version, s.Level, s.Edition)
}
