package networking

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"prjfree/client/crypt"
	sqlm "prjfree/client/data"
	"prjfree/client/models"
	"strings"
)

type Client struct {
	M        models.Commutator
	Cl       string
	conn_num int
	From_ip  string
}

type SQLBlock struct {
	Id    int
	Hash  string
	Date  string
	Num   int
	Topic string
	Data  string
}

var clientqueue []Client

func AddClient() {
	if len(clientqueue) > 0 {
		client := clientqueue[0]
		clientqueue = clientqueue[1:]
		cls := models.Comms[client.From_ip]
		cls.AddClient(client.Cl)
		models.Comms[client.From_ip] = cls
	}
}

func WaitForMessages(conn *Conn) {
	scanr := bufio.NewScanner(conn.reader)
	for {
		scanned := scanr.Scan()
		if !scanned {
			if err := scanr.Err(); err != nil {
				fmt.Printf("%v(%v)\n", err, conn.RemoteAddr())
				return
			}
			break
		}
		resp := strings.Split(scanr.Text(), ";;;")
		if len(resp) >= 3 && resp[0] == "[MESSAGE]" && resp[2] == "{DISCOVER}" {
			SendMessage(conn, "{MESSAGE};;;"+resp[1]+";;;"+models.GetCommutatorsToString()+"\n")
		} else if len(resp) >= 3 && resp[0] == "[CLIENTSRESULT]" {
			from_ip := resp[1]
			clients := strings.Split(resp[2], ",")
			m, ok := models.Comms[from_ip]
			if ok {
				for _, client := range clients {
					if client != "" {
						cl := Client{
							M:       m,
							From_ip: from_ip,
							Cl:      client,
						}
						clientqueue = append(clientqueue, cl)
					}
				}
			}
		} else if len(resp) >= 2 && resp[0] == "[DISCOVERRESULT]" {
			commutators := strings.Split(resp[1], ",")
			for _, commutator := range commutators {
				models.Comms[commutator] = models.Commutator{
					IP:   strings.Split(commutator, ":")[0],
					Port: strings.Split(commutator, ":")[1],
				}
				ConnectToCommutator(models.Comms[commutator])
			}
		} else if resp[0] == "[MESSAGE]" {
			from := resp[1]
			_, ok := Sessions[from]
			params := strings.Split(resp[2], ";")
			if params[0] == "[STARTSESSION]" && !ok {
				publicKey := crypt.DenormalizeText(params[1])
				fmt.Println(from)
				session := Session{
					PubKey: crypt.BytesToPublicKey([]byte(publicKey)),
					Conn:   conn,
					Code:   from,
				}
				Sessions[from] = session
				StartSession(conn, from)
			} else if params[0] == "[ENCMESSAGE]" {
				data, _ := b64.StdEncoding.DecodeString(params[1])
				if ok {
					decoded := string(crypt.DecryptWithPrivateKey(data, crypt.PrivateKey))
					decoded = strings.Trim(decoded, " ")
					decoded_data := strings.Split(decoded, ";")
					if decoded_data[0] == "[BLOCK]" {
						dat := decoded_data[2]
						hash := decoded_data[1]
						datetime := decoded_data[4]
						num := decoded_data[3]
						topic := strings.Split(dat, "::")[0]
						_, err := sqlm.DB.Exec("INSERT INTO blocks VALUES(NULL, $1, $2, $3, $4, $5);", hash, datetime, num, topic, strings.Split(dat, "::")[1])
						if err != nil {
							fmt.Printf("Error: %v\n", err.Error())
						}
						fmt.Println("Command block")
					} else if decoded_data[0] == "[FIND]" {
						topic := decoded_data[1]
						res, err := sqlm.DB.Query("SELECT * FROM blocks WHERE topic=$1", topic)
						defer res.Close()
						if err != nil {
							fmt.Printf("Error during find(%v)\n", err.Error())
						} else {

							for res.Next() {
								b := SQLBlock{}
								err := res.Scan(&b.Id, &b.Hash, &b.Date, &b.Num, &b.Topic, &b.Data)
								if err != nil {
									fmt.Printf("Error during sql query: %v\n", err.Error())
								} else {
									SendFindResult(b, conn, from)
								}
							}
						}
					} else if decoded_data[0] == "[FOUND]" {
						hash := decoded_data[1]
						date := decoded_data[2]
						num := decoded_data[3]
						topic := decoded_data[4]
						dat := decoded_data[5]
						_, err := sqlm.DB.Exec("INSERT INTO blocks VALUES(NULL, $1, $2, $3, $4, $5);", hash, date, num, topic, dat)
						if err != nil {
							fmt.Printf("Error: %v\n", err.Error())
						}
					}
					fmt.Printf("Yea we got decrypted message: %v\n", decoded)
				}
			}
			fmt.Println(scanr.Text())
		}
	}
}
