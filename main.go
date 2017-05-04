package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	serial "go.bug.st/serial.v1"

	rice "github.com/GeertJohan/go.rice"
	"github.com/ghodss/yaml"
	"github.com/hlidotbe/macropad/auxilium"
	"github.com/hlidotbe/macropad/pad"
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

type actionConfig struct {
	Type          string   `json:"type"`
	ID            int      `json:"id"`             // For Track actions
	Label         string   `json:"label"`          // For Track actions
	Profile       string   `json:"profile"`        // For Track actions
	DisplayOutput bool     `json:"display_output"` // For Macro actions
	Args          []string `json:"args"`           // For Type and Macro actions
	Duration      int      `json:"duration"`       // For Pomodoro actions
}

var auxiliumClient *auxilium.Client
var config *map[string]*actionConfig
var orch *pad.Orchestrator

func main() {
	config = loadConfig()

	go setupHTTP()

	port := openPort()

	auxiliumClient = auxilium.NewClient(nil, os.Getenv("AUXILIUM_TOKEN"), "https://track.epic.net/api")

	orch = pad.NewOchestrator(port)
	setupKeys()

	orch.Run()
}

func setupHTTP() {
	http.HandleFunc("/keys", handleKeys)
	http.Handle("/", http.FileServer(rice.MustFindBox("templates").HTTPBox()))
	log.Fatal(http.ListenAndServe(":6276", nil))
}

func loadConfig() *map[string]*actionConfig {
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
	cfg := new(map[string]*actionConfig)
	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

//*
func openPort() serial.Port {
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
	return port
}

//*/
/*
func openPort() io.ReadWriter {
	return &rw{in: os.Stdin, out: os.Stdout}
}

//*/

func setupKeys() {
	for key, ac := range *config {
		switch ac.Type {
		case "Track":
			orch.RegisterAction(key, pad.NewActionTrack(key, orch.Com, auxiliumClient, ac.Label, ac.ID, ac.Profile))
			break
		case "Type":
			orch.RegisterAction(key, pad.NewActionType(key, orch.Com, ac.Args...))
			break
		case "Macro":
			orch.RegisterAction(key, pad.NewActionMacro(key, orch.Com, ac.DisplayOutput, ac.Args...))
			break
		case "Pomodoro":
			orch.RegisterAction(key, pad.NewActionPomodoro(key, orch.Com, time.Duration(ac.Duration)*time.Minute))
			break
		}
		log.Printf("%v: %v\n", key, ac)
	}
}

func handleKeys(response http.ResponseWriter, request *http.Request) {
	k := request.URL.Query()["k"][0]
	if request.Method == "GET" {
		ac := (*config)[k]
		if ac == nil {
			response.WriteHeader(404)
			return
		}
		bytes, _ := json.Marshal(ac)
		response.Write(bytes)
	} else {
		ac := (*config)[k]
		if ac != nil {
			// TODO: Unregister action
		}
		(*config)[k] = nil
	}
}
