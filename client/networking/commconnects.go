package networking

import (
	"bufio"
	"crypto/rsa"
	b64 "encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"prjfree/client/crypt"
	sqlm "prjfree/client/data"
	"prjfree/client/models"
	"regexp"
	"strconv"
	"time"
)

type Session struct {
	PubKey *rsa.PublicKey
	Conn   *Conn
	Code   string
}

type Conn struct {
	net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	commut models.Commutator
}

var Conns []*Conn

var Tasks []models.Block

var Sessions map[string]Session = make(map[string]Session, 0)

func AddTask(block models.Block) {
	Tasks = append(Tasks, block)
}

func PopTask() models.Block {
	task := Tasks[0]
	Tasks = Tasks[1:]
	return task
}

func SetInterval(f func(), mills int, async bool) chan bool {
	interval := time.Duration(mills) * time.Millisecond

	ticker := time.NewTicker(interval)
	clear := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if async {
					go f()
				} else {
					f()
				}
			case <-clear:
				ticker.Stop()
				return
			}
		}
	}()
	return clear
}

func SendClientsDiscover(conn *Conn, comm models.Commutator) {
	SendMessage(conn, "{CLIENTS};;;"+comm.IP+":"+comm.Port+"\n")
}

func SendEncryptedMessage(conn *Conn, client string, msg string) {
	_, ok := Sessions[client]
	if ok {
		message := "{MESSAGE};;;" + client + ";;;[ENCMESSAGE];"
		fmt.Printf("Encrypted data: %v\n", []byte(msg))
		encryptedData := b64.StdEncoding.EncodeToString(crypt.EncryptWithPublicKey([]byte(msg), Sessions[client].PubKey))
		message += encryptedData + "\n"
		SendMessage(conn, message)
	}
}

func GenCode(length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func ConnectToCommutator(comm models.Commutator) {
	fmt.Printf("Connecting to %v\n", comm.IP+":"+comm.Port)
	nconn, err := net.Dial("tcp", comm.IP+":"+comm.Port)
	if err == nil {
		conn := &Conn{
			Conn:   nconn,
			commut: comm,
		}
		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)
		conn.reader = r
		conn.writer = w
		Conns = append(Conns, conn)
		go WaitForMessages(conn)
		SendClientsDiscover(conn, comm)
		SetInterval(func() {
			SendClientsDiscover(conn, comm)
		}, 10000, true)
	} else {
		fmt.Printf("Error during connection to commutator(%v)\n", err.Error())
	}
}

func SendFindResult(block SQLBlock, conn *Conn, client string) {
	msg := "[FOUND];" + block.Hash + ";" + block.Date + ";" + strconv.Itoa(block.Num) + ";" + block.Topic + ";" + block.Data
	SendEncryptedMessage(conn, client, msg)
}

func CommsConnect(comms map[string]models.Commutator) {
	for _, comm := range comms {
		ConnectToCommutator(comm)
	}
}

func SendMessage(conn *Conn, message string) {
	conn.writer.WriteString(message)
	conn.writer.Flush()
}

func StartTasks() {
	SetInterval(func() { DoTask() }, models.TASK_TIME, true)
}

func DiscoverCommutators() {
	for _, conn := range Conns {
		fmt.Printf("Sending 'discover' for %v\n", conn.commut.IP+":"+conn.commut.Port)
		SendMessage(conn, "{MESSAGEFORALL};;;{DISCOVER}\n")
	}
}

func getRand(a int, b int) int {
	if b-a == 0 {
		return a
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(b-a) + a
}

func getRandSession() Session {
	rndnum := getRand(0, len(Sessions))
	fmt.Printf("Randnum: %v, len: %v\n", rndnum, len(Sessions))
	i := 0
	for _, v := range Sessions {
		if i == rndnum {
			return v
		}
		i++
	}
	return Session{}
}

func getPossibleSessionCount() int {
	res := 0
	for _, comm := range models.Comms {
		for _, _ = range comm.Clients {
			res++
		}
	}
	return res
}

func genRange(min, max int) []int {
	res := make([]int, max-min+1)
	for i := range res {
		res[i] = min + i
	}
	return res
}

func getAllClients() []Client {
	res := make([]Client, 0)
	j := 0
	for _, comm := range models.Comms {
		for _, cl := range comm.Clients {
			res = append(res, Client{
				M:        comm,
				Cl:       cl,
				conn_num: j,
			})
		}
		j++
	}
	return res
}

func remove(a []int, ind int) []int {
	a[ind], a[len(a)-1] = a[len(a)-1], a[ind]
	return a[:len(a)-1]
}

func StartExchange() {
	fmt.Println("Starting exchange...")
	SetInterval(func() { Exchange() }, 20000, true)
}

func FormData(sqlb SQLBlock) models.Block {
	var res models.Block
	res.Data = []byte(sqlb.Hash + ";" + sqlb.Topic + "::" + string(sqlb.Data) + ";" + strconv.Itoa(sqlb.Num) + ";" + sqlb.Date)
	return res
}

func Exchange() {
	res, err := sqlm.DB.Query("SELECT * FROM blocks ORDER BY RANDOM() LIMIT 1;")
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
				block := FormData(b)
				AddTask(block)
			}
		}
	}
}

func DisplayComms() {
	for k, v := range models.Comms {
		fmt.Println(k + ":")
		for _, cl := range v.Clients {
			fmt.Printf("    -%v ", cl)
			_, ok := Sessions[cl]
			if ok {
				fmt.Printf("[SESSION]\n")
			} else {
				fmt.Printf("\n")
			}
		}
	}
}

func GenSessions(count int) {
	all_clients := getAllClients()
	conn_range := genRange(0, len(all_clients)-1)
	for i := 0; i < count; i++ {
		if len(conn_range) == 0 {
			break
		}
		num := getRand(0, len(conn_range))
		ind := conn_range[num]
		conn_range = remove(conn_range, num)
		StartSession(Conns[all_clients[ind].conn_num], all_clients[ind].Cl)
	}
}

func StartSession(conn *Conn, client string) {
	message := "{MESSAGE};;;" + client + ";;;[STARTSESSION];"
	message += crypt.NormalizeText(string(crypt.PublicKeyToBytes(crypt.PublicKey)))
	message += ";"
	message += "\n"
	fmt.Printf("Message: %v\n%v", message, conn)
	SendMessage(conn, message)
}

func Unspace(s string) string {
	reg, err := regexp.Compile("[^a-zA-Zа-яА-Я ]+")
	if err != nil {
		return ""
	}
	return reg.ReplaceAllString(s, "")
}

func DoTask() {
	if len(Tasks) > 0 {
		task := Tasks[0]
		Tasks = Tasks[1:]
		for i := 0; i < models.STABILITY; i++ {
			session := getRandSession()
			fmt.Printf("Task code: %v\n", "[BLOCK];"+string(task.Data))
			SendEncryptedMessage(session.Conn, session.Code, "[BLOCK];"+string(task.Data))
			//client := getRandClient(conn.commut)
			//StartSession(conn, client)
		}
	}
}

func SendData(d []byte, topic string) {
	blocks := crypt.DataToBlocks(d)
	dt := time.Now()
	date := dt.Format("2006-01-02 15:04:05")
	hsh := GenCode(32)
	for i := 0; i < len(blocks); i++ {
		dt2 := hsh + ";" + topic + "::" + string(blocks[i].Data) + ";" + strconv.Itoa(i) + ";" + date
		AddTask(models.Block{
			Data: []byte(dt2),
		})
	}
}

func Find(topic string) {
	fmt.Println("Sessions count:", len(Sessions))
	for _, sess := range Sessions {
		SendEncryptedMessage(sess.Conn, sess.Code, "[FIND];"+topic)
	}
}
