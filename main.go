package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/hlidotbe/macropad/auxilium"
	"github.com/hlidotbe/macropad/pad"
	"github.com/smallfish/simpleyaml"

	serial "go.bug.st/serial.v1"
)

type rw struct {
	in  io.Reader
	out io.Writer
}

func (s *rw) Read(p []byte) (n int, err error) {
	n, err = s.in.Read(p)
	return n, err
}

func (s *rw) Write(p []byte) (n int, err error) {
	n, err = s.out.Write(p)
	return n, err
}

var auxiliumClient *auxilium.Client
var config *simpleyaml.Yaml
var orch *pad.Orchestrator

func main() {
	config = loadConfig()

	go setupHTTP()

	port := openPort()

	auxiliumClient = auxilium.NewClient(nil, os.Getenv("AUXILIUM_TOKEN"), "https://track.epic.net/api")

	orch = pad.NewOchestrator(*port)
	setupKeys()

	orch.Run()
}

func setupHTTP() {
	http.HandleFunc("/keys", handleKeys)
	http.Handle("/", http.FileServer(rice.MustFindBox("templates").HTTPBox()))
	log.Fatal(http.ListenAndServe(":6276", nil))
}

func loadConfig() *simpleyaml.Yaml {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	configFile := path.Join(u.HomeDir, ".macropad.yml")
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	s, _ := file.Stat()
	buf := make([]byte, s.Size())
	if _, err = file.Read(buf); err != nil {
		log.Fatal(err)
	}
	y, err := simpleyaml.NewYaml(buf)
	if err != nil {
		log.Fatal(err)
	}
	return y
}

func openPort() *serial.Port {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	var portName string
	for _, p := range ports {
		if strings.HasPrefix(p, "/dev/tty.usbmodem") {
			portName = p
			break
		}
	}
	fmt.Printf("Found port: %v\n", portName)
	mode := &serial.Mode{}
	port, err := serial.Open(portName, mode)
	if err != nil {
		log.Fatal(err)
	}
	return &port
}

func setupKeys() {
	m, _ := config.Map()
	for k, v := range m {
		key := k.(string)
		prop := v.(map[interface{}]interface{})
		typ := prop["type"].(string)
		switch typ {
		case "Track":
			orch.RegisterAction(key, pad.NewActionTrack(key, orch.Com, auxiliumClient, prop["label"].(string), prop["id"].(int), prop["profile"].(string)))
			break
		case "Type":
			argsi := prop["args"].([]interface{})
			args := make([]string, len(argsi))
			for i, v := range argsi {
				args[i] = v.(string)
			}
			orch.RegisterAction(key, pad.NewActionType(key, orch.Com, args...))
			break
		case "Macro":
			argsi := prop["args"].([]interface{})
			args := make([]string, len(argsi))
			for i, v := range argsi {
				args[i] = v.(string)
			}
			orch.RegisterAction(key, pad.NewActionMacro(key, orch.Com, prop["display_output"].(bool), args...))
			break
		case "Pomodoro":
			orch.RegisterAction(key, pad.NewActionPomodoro(key, orch.Com, time.Duration(prop["duration"].(int))*time.Minute))
			break
		}
		log.Printf("%v: %v\n", k, v)
	}
}

func handleKeys(response http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {

	} else {
	}
}
