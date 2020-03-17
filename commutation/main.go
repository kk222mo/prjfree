package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

type Conn struct {
	net.Conn
	IdleTimeout time.Duration
	code        string
	writer      *bufio.Writer
	reader      *bufio.Reader
}

var clients []*Conn

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (c *Conn) Write(p []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Write(p)
}

func (c *Conn) Read(b []byte) (int, error) {
	c.updateDeadline()
	return c.Conn.Read(b)
}

func (c *Conn) updateDeadline() {
	idleDeadline := time.Now().Add(c.IdleTimeout)
	c.Conn.SetDeadline(idleDeadline)
}

func getClients(cl string) string {
	var res string

	for _, elem := range clients {
		if elem.code != cl {
			res += elem.code + ","
		}
	}
	return res
}

func genCode(length int) string {
	charset := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func sendMessageByCode(from string, to string, message string) {
	for _, cl := range clients {
		if cl.code == to {
			cl.writer.WriteString("[MESSAGE];;;" + from + ";;;" + message)
			cl.writer.Flush()
			break
		}
	}
}

func handle(conn *Conn) {
	fmt.Printf("%v connected\n", conn.RemoteAddr())
	defer func() {
		for ind, cl := range clients {
			if cl == conn {
				clients[ind] = clients[len(clients)-1]
				clients[len(clients)-1] = nil
				clients = clients[:len(clients)-1]
				break
			}
		}
		fmt.Printf("%v disconnected\n", conn.RemoteAddr())
		conn.Close()
	}()

	clients = append(clients, conn)
	conn.code = genCode(30)

	conn.SetDeadline(time.Now().Add(conn.IdleTimeout))
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	conn.reader = r
	conn.writer = w
	scanr := bufio.NewScanner(r)
	for {
		scanned := scanr.Scan()
		if !scanned {
			if err := scanr.Err(); err != nil {
				fmt.Printf("%v(%v)\n", err, conn.RemoteAddr())
				return
			}
			break
		}
		command := strings.Split(scanr.Text(), ";;;")
		if command[0] == "{CLIENTS}" {
			w.WriteString("[CLIENTSRESULT];;;" + command[1] + ";;;" + getClients(conn.code) + "\n")
			w.Flush()
		} else if command[0] == "{MESSAGE}" && len(command) >= 3 {
			fmt.Printf("Sending message from %v to %v\n", conn.code, command[1])
			sendMessageByCode(conn.code, command[1], command[2]+"\n")
		} else if command[0] == "{MESSAGEFORALL}" && len(command) >= 2 {
			fmt.Printf("Sending message from %v to all\n", conn.code)
			for _, cl := range clients {
				if cl.code != conn.code {
					sendMessageByCode(conn.code, cl.code, command[1]+"\n")
				}
			}
		}
	}
}

func ListenTCP(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer listener.Close()

	for {
		nconn, err := listener.Accept()
		if err != nil {
			continue
		}

		conn := &Conn{
			Conn:        nconn,
			IdleTimeout: 10,
		}
		conn.IdleTimeout, _ = time.ParseDuration("24h")

		go handle(conn)
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8685", "port")
	flag.Parse()
	fmt.Println("Listening for connections...")
	ListenTCP(":" + port)
}
