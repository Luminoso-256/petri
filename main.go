package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

var config Config

const version = "Petri 1.0.2"

func handleConnection(c net.Conn) {
	fmt.Printf("Connection Made\n")
	var requestBytes []byte
	var clen uint16
	path := ""
	clen = 0
	gotclen := false
	reader := bufio.NewReader(c)
	for {
		b, _ := reader.ReadByte()
		requestBytes = append(requestBytes, b)
		//read our "header" so we know when to stop
		if len(requestBytes) >= 2 && !gotclen {
			clb := requestBytes[:2]
			buf := bytes.NewBuffer(clb)
			binary.Read(buf, binary.LittleEndian, &clen)
			gotclen = true
			fmt.Printf("Clen = %v", clen)
		}
		//if the len of our byte array = the content len, it's time to stop recieving
		if (len(requestBytes) >= int(clen)+2) && gotclen {
			break
		}
	}
	path = string(requestBytes[2:])
	fmt.Printf("Request recieved for path %s (total bytes read: %v / content len: %v) \n", path, len(requestBytes), clen)
	if strings.Contains(path, "..") {
		fmt.Println("Request contains .. - closing prematurely due to assumed security breach attempt")
		c.Close()
		return
	}

	fullpath := "data/srv" + path
	info, err := os.Stat(fullpath)
	if err != nil {
		//we don't have what you're looking for. Sorry!
		c.Write([]byte{0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
		c.Close()
		return
	}
	if info.IsDir() {
		//check if we want directory listings. if not, return 0x22
		if config.ListDirs {
			files, err := ioutil.ReadDir(fullpath)
			if err != nil {
				c.Write([]byte{0x23, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
			}
			output := "Directory Listing of " + path + "\n"
			for _, f := range files {
				output += "=> piper://" + config.Hostname + path + f.Name() + " " + f.Name() + "\n"
			}
			output += "> " + version + "\n"
			var response []byte
			response = append(response, 0x01)
			conb := []byte(output)
			lenb := make([]byte, 8)
			binary.LittleEndian.PutUint64(lenb, uint64(len(conb)))
			response = append(response, lenb...)
			response = append(response, conb...)
			c.Write(response)
		} else {
			c.Write([]byte{0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
		}
	} else {
		//get the file! (and parse extension)
		datab, err := ioutil.ReadFile(fullpath)
		if err != nil {
			c.Write([]byte{0x23, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
		}
		var response []byte
		switch strings.Split(path, ".")[len(strings.Split(path, "."))-1] {
		case "txt":
			response = append(response, 0x00)
			break
		case "gmi":
			response = append(response, 0x01)
			break
		case "ascii":
			response = append(response, 0x02)
			break
		default:
			response = append(response, 0x10)
			break
		}
		lenb := make([]byte, 8)
		binary.LittleEndian.PutUint64(lenb, uint64(len(datab)))
		response = append(response, lenb...)
		response = append(response, datab...)
		c.Write(response)
	}
	c.Close()
}

func main() {
	fmt.Println("Petri: The fast, simple, and flexible Piper Webserver")
	datab, _ := ioutil.ReadFile("data/config.json")
	json.Unmarshal(datab, &config)
	l, err := net.Listen("tcp4", fmt.Sprintf(":%v", (config.Port)))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
