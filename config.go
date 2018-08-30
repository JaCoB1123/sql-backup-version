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

type stringer string

func (s stringer) String() string {
	return string(s)
}

type serverList []server
type databaseList []stringer

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

type selectable interface {
	getLength() int
	getElement(int) fmt.Stringer
}

func (sl serverList) getLength() int {
	return len(sl)
}

func (sl serverList) getElement(i int) fmt.Stringer {
	return &sl[i]
}

func (dl databaseList) getLength() int {
	return len(dl)
}

func (dl databaseList) getElement(i int) fmt.Stringer {
	return &dl[i]
}
