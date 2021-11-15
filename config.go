package main

type Config struct {
	Port     int    `json:"port"`
	ListDirs bool   `json:"listdirectories"`
	Hostname string `json:"hostname"`
}
