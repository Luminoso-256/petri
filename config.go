package main

type Config struct {
	Port     int    `json:"port"`
	ListDirs bool   `json:"listdirectories"`
	Hostname string `json:"hostname"`
}

type FiletypeConfig struct {
	TextUtf8  []string `json:"text:utf8"`
	TextGem   []string `json:"text:gem"`
	TextASCII []string `json:"text:ascii"`
	SrvRedir  []string `json:"srv:redir"`
}
