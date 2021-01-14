package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

type config struct {
	LogFilePath string `json:"log_file_path"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	HTTP        *struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"http"`
}

func defaultConfig() *config {
	return new(config)
}

func readConfigFile(filename string) (*config, error) {
	filename = filepath.FromSlash(filename) // cope with OS dependent oddities

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := new(config)

	err = json.Unmarshal(bs, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
