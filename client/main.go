package main

import (
	"bufio"
	"fmt"
	"os"
	enc "prjfree/client/crypt"
	"prjfree/client/data"
	"prjfree/client/models"
	"prjfree/client/networking"
	"prjfree/client/webgui"
	"strings"
)

func main() {
	data.LoadDB("data.db")
	networking.SetInterval(func() { networking.AddClient() }, 1000, true)
	enc.PrivateKey, enc.PublicKey = enc.GenKeyPair(models.KEY_LEN)
	models.LoadCommutatorsFromFile("commutators.txt")
	networking.CommsConnect(models.Comms)
	networking.DiscoverCommutators()
	reader := bufio.NewReader(os.Stdin)
	networking.StartTasks()
	for {
		fmt.Printf("Command: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		pars := strings.Split(text, ";")
		if text == "INFO" {
			networking.DisplayComms()
			fmt.Println(webgui.GetClients())
		} else if pars[0] == "SESSIONS" {
			networking.GenSessions(models.CONN_COUNT)
		} else if pars[0] == "PLAINENC" {
			networking.SendEncryptedMessage(networking.Conns[0], pars[1], pars[2])
		} else if pars[0] == "ENC" {
			networking.SendData([]byte(pars[1]), pars[2])
		} else if pars[0] == "FIND" {
			topic := pars[1]
			networking.Find(topic)
		} else if pars[0] == "EXCHANGE" {
			fmt.Println("Starting exchange...")
			networking.StartExchange()
		} else if pars[0] == "GUI" {
			webgui.StartListening()
			err := webgui.OpenBrowser()
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}
