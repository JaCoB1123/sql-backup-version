package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type fileType int

const (
	Public fileType = iota + 1
	Local
)

type fileshare struct {
	Path string
	Type fileType
}

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

// Configuration saves the configuration
type configuration struct {
	Servers []server
	Files   []fileshare
}

// GetConfig returns the saved configuration
func getConfig() (*configuration, error) {
	serverConfig, err := ioutil.ReadFile("./config/servers.json")
	if err != nil {
		return nil, err
	}

	var configuration configuration
	var servers []server
	err = json.Unmarshal(serverConfig, &servers)
	if err != nil {
		return nil, err
	}

	configuration.Servers = servers

	fileConfig, err := ioutil.ReadFile("./config/fileshares.json")
	if err != nil {
		return nil, err
	}

	var files []fileshare
	err = json.Unmarshal(fileConfig, &files)
	if err != nil {
		return nil, err
	}

	configuration.Files = files

	return &configuration, nil
}

func (s server) String() string {
	return fmt.Sprintf("%s\\%s\nVersion: %s (%s)\nEdition: %s", s.Host, s.Instance, s.Version, s.Level, s.Edition)
}
