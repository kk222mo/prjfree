package webgui

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os/exec"
	sqlm "prjfree/client/data"
	"prjfree/client/models"
	"prjfree/client/networking"
	"runtime"
	"strings"
)

type GuiBlock struct {
	dat string
}

func GetClients() string {
	var cmms map[string]models.Commutator = make(map[string]models.Commutator, 0)
	for k, v := range models.Comms {
		cmms[k] = v
	}
	b, _ := json.Marshal(cmms)
	return string(b)
}

func GetTasks() string {
	t, _ := json.Marshal(networking.Tasks)
	return string(t)
}

func listComms(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, GetClients())
}

func listTasks(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, GetTasks())
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	tmpl.Execute(w, nil)
}

func OpenBrowser() error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, "http://127.0.0.1:"+models.GUI_PORT)
	return exec.Command(cmd, args...).Start()
}

func startSessions(w http.ResponseWriter, r *http.Request) {
	networking.GenSessions(models.CONN_COUNT)
}

func sessionAvailable(w http.ResponseWriter, r *http.Request) {
	client := strings.Split(r.URL.RawQuery, "=")[1]
	_, ok := networking.Sessions[client]
	if !ok {
		fmt.Fprint(w, "false;"+client)
		return
	}
	fmt.Fprint(w, "true;"+client)
}
func sendMsg(w http.ResponseWriter, r *http.Request) {
	msg, _ := url.QueryUnescape(strings.Split(r.URL.RawQuery, "=")[1])
	networking.SendData([]byte(msg), "Hello")
}

func DisplayResults() string {
	res, err := sqlm.DB.Query("SELECT * FROM blocks ORDER BY date,num")
	resblocks := make(map[string]map[string]string)
	defer res.Close()
	if err != nil {
		fmt.Printf("Error during find(%v)\n", err.Error())
	} else {

		for res.Next() {
			b := networking.SQLBlock{}
			_ = res.Scan(&b.Id, &b.Hash, &b.Date, &b.Num, &b.Topic, &b.Data)
			_, ok := resblocks[b.Date]
			if !ok {
				resblocks[b.Date] = map[string]string{b.Hash: b.Data}
			} else {
				resblocks[b.Date][b.Hash] += b.Data
			}
		}
	}
	//fmt.Println(resblocks)
	js, err := json.Marshal(resblocks)
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(js)
}

func listData(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, DisplayResults())
}

func startExchanging(w http.ResponseWriter, r *http.Request) {
	networking.StartExchange()
}

func StartListening() {
	http.HandleFunc("/", index)
	http.HandleFunc("/comms", listComms)
	http.HandleFunc("/tasks", listTasks)
	http.HandleFunc("/blocks", listData)
	http.HandleFunc("/sessions", startSessions)
	http.HandleFunc("/exchange", startExchanging)
	http.HandleFunc("/asession", sessionAvailable)
	http.HandleFunc("/send-msg", sendMsg)
	go http.ListenAndServe(":"+models.GUI_PORT, nil)
}
