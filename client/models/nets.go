package models

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Commutator struct {
	IP      string
	Port    string
	Clients []string
}

func check(err error) {
	if err != nil {
		fmt.Printf("Error: %v", err.Error())
		os.Exit(0)
	}
}

var Comms map[string]Commutator

func contains(clients []string, client string) bool {
	for _, cl := range clients {
		if cl == client {
			return true
		}
	}
	return false
}

func (comm *Commutator) AddClient(client string) {
	if !contains(comm.Clients, client) {
		comm.Clients = append(comm.Clients, client)
	}
}

func LoadCommutatorsFromFile(path string) {
	Comms = make(map[string]Commutator)
	data, err := ioutil.ReadFile(path)
	check(err)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line != "" {
			pars := strings.Split(line, ":")
			if len(pars) == 2 {
				comm := Commutator{
					IP:   pars[0],
					Port: pars[1],
				}
				Comms[pars[0]+":"+pars[1]] = comm
			}
		}
	}
}

func GetCommutatorsToString() string {
	var res string
	for _, el := range Comms {
		res += el.IP + ":" + el.Port + ","
	}
	return res
}
