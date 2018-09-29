package main

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Slack struct {
	Token   string `toml:"token"`
	Channel string `toml:"channel"`
	User    string `toml:"user"`
}

type Jira struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	Endpoint string `toml:"endpoint"`
	ServerID string `toml:"server-id"`
	Server   string `toml:"server"`
	Project  string `toml:"project"`
	OnCall   string `toml:"oncall"`
}

type Member struct {
	Name   string `json:"name"`
	Github string `json:"github"`
	Email  string `json:"email"`
}

type Team struct {
	Name    string   `json:"name"`
	Members []Member `json:"members"`
}

type Github struct {
	Token string   `json:"token"`
	Repos []string `json:"repos"`
}

type Confluence struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	Endpoint string `toml:"endpoint"`

	Space      string `toml:"space"`
	WeeklyPath string `toml:"weekly-path"`
}

type Config struct {
	Slack      Slack      `toml:"slack"`
	Jira       Jira       `toml:"jira"`
	Confluence Confluence `toml:"confluence"`
	Github     Github     `toml:"github"`
	Teams      []Team     `toml:"teams"`
}

// NewConfigFromFile creates the configuration from file
func NewConfigFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c := new(Config)
	if err = toml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	return c, nil
}
