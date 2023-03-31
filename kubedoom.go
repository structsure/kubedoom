package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// hash generates a hash from a given input string.
func hash(input string) int32 {
	var hash int32 = 5381
	for _, char := range input {
		hash = ((hash << 5) + hash + int32(char))
	}
	if hash < 0 {
		hash = 0 - hash
	}
	return hash
}

// runCommand executes the given command and logs if there's an error.
func runCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", cmd)
	}
}

// outputCommand executes the given command and returns its output as a string.
func outputCommand(cmd *exec.Cmd) string {
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", cmd)
	}
	return string(output)
}

// startCommand starts the given command without waiting for it to complete.
func startCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Start()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", cmd)
	}
}

type Mode interface {
	getEntities() []string
	deleteEntity(string)
}

type podmode struct{}

func (m podmode) getEntities() []string {
	var args []string
	if namespace, exists := os.LookupEnv("NAMESPACE"); exists {
		args = []string{"kubectl", "get", "pods", "--namespace", namespace, "-o", "go-template", "--template={{range .items}}{{.metadata.namespace}}/{{.metadata.name}} {{end}}"}
	} else {
		args = []string{"kubectl", "get", "pods", "-A", "-o", "go-template", "--template={{range .items}}{{.metadata.namespace}}/{{.metadata.name}} {{end}}"}
	}
	output := outputCommand(exec.Command(args[0], args[1:]...))
	outputstr := strings.TrimSpace(output)
	pods := strings.Split(outputstr, " ")
	return pods
}

func (m podmode) deleteEntity(entity string) {
	log.Printf("Pod to kill: %v", entity)
	podparts := strings.Split(entity, "/")
	cmd := exec.Command("/usr/bin/kubectl", "delete", "pod", "-n", podparts[0], podparts[1])
	go cmd.Run()
}

type nsmode struct{}

func (m nsmode) getEntities() []string {
	args := []string{"kubectl", "get", "namespaces", "-o", "go-template", "--template={{range .items}}{{.metadata.name}} {{end}}"}
	output := outputCommand(exec.Command(args[0], args[1:]...))
	outputstr := strings.TrimSpace(output)
	namespaces := strings.Split(outputstr, " ")
	return namespaces
}

func (m nsmode) deleteEntity(entity string) {
	log.Printf("Namespace to kill: %v", entity)
	cmd := exec.Command("/usr/bin/kubectl", "delete", "namespace", entity)
	go cmd.Run()
}

func socketLoop(listener net.Listener, mode Mode) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		stop := false
		for !stop {
			bytes := make([]byte, 40960)
			n, err := conn.Read(bytes)
			if err != nil {
				stop = true
			}
			bytes = bytes[0:n]
			strbytes := strings.TrimSpace(string(bytes))
			entities := mode.getEntities()
			if strbytes == "list" {
				for _, entity := range entities {
					padding := strings.Repeat("\n", 255-len(entity))
					_, err = conn.Write([]byte(entity + padding))
					if err != nil {
						log.Fatal("Could not write to socket file")
					}
				}
				conn.Close()
				stop = true
			} else if strings.HasPrefix(strbytes, "kill ") {
				parts := strings.Split(strbytes, " ")
				killhash, err := strconv.ParseInt(parts[1], 10, 32)
				if err != nil {
					log.Fatal("Could not parse kill hash")
				}
				for _, entity := range entities {
					if hash(entity) == int32(killhash) {
						mode.deleteEntity(entity)
						break
					}
				}
				conn.Close()
				stop = true
			}
		}
	}
}

func main() {
	var modeFlag string
	flag.StringVar(&modeFlag, "mode", "pods", "What to kill pods|namespaces")

	flag.Parse()

	var mode Mode
	switch modeFlag {
	case "pods":
		mode = podmode{}
	case "namespaces":
		mode = nsmode{}
	default:
		log.Fatalf("Mode should be pods or namespaces")
	}

	listener, err := net.Listen("unix", "/dockerdoom.socket")
	if err != nil {
		log.Fatalf("Could not create socket file")
	}

	log.Print("Create virtual display")
	runCommand(exec.Command("/usr/bin/Xvfb", ":99", "-ac", "-screen", "0", "640x480x24"))
	time.Sleep(time.Duration(2) * time.Second)
	startCommand(exec.Command("x11vnc", "-geometry", "640x480", "-forever", "-usepw", "-display", ":99"))
	log.Print("You can now connect to it with a VNC viewer at port 5900")

	log.Print("Trying to start DOOM ...")
	startCommand(exec.Command("/usr/bin/env", "DISPLAY=:99", "/usr/local/games/psdoom", "-warp", "-E1M1", "-skill", "1", "-nomouse"))
	socketLoop(listener, mode)
}
