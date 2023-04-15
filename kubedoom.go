package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func hash(input string) int32 {
	var hash int32
	hash = 5381
	for _, char := range input {
		hash = ((hash << 5) + hash + int32(char))
	}
	if hash < 0 {
		hash = 0 - hash
	}
	return hash
}

func runCmd(cmdstring string) {
	parts := strings.Split(cmdstring, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", cmdstring)
	}
}

func outputCmd(argv []string) string {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", argv)
	}
	return string(output)
}

func startCmd(cmdstring string) {
	parts := strings.Split(cmdstring, " ")
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Start()
	if err != nil {
		log.Fatalf("The following command failed: \"%v\"\n", cmdstring)
	}
}

func NewClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(config)
}

func formatEntityName(pod v1.Pod) string {
    return fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())
}

func socketLoop(listener net.Listener, podsListChan <-chan *v1.PodList) {
	clientset := NewClientSet()
	// current pod name is hostname
	hostname, err := os.Hostname()
    if err != nil {
        panic(err)
    }
	log.Printf("Starting socket loop for %f", hostname)

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
			podsList := <- podsListChan
			if strbytes == "list" {
				log.Printf("Sending entity list")
				for _, pod := range podsList.Items {
					// filter out certain pods
					if strings.HasPrefix(pod.Name, "kubedoom") {
						if pod.Name == hostname {
							log.Printf("Filtering out %v", pod.Name)
							continue
						}
					}
					entity := formatEntityName(pod)
					padding := strings.Repeat("\n", 255-len(entity))
					_, err = conn.Write([]byte(entity + padding))
					if err != nil {
						log.Fatal("Could not write to socker file")
					}
				}
				conn.Close()
				stop = true
				log.Printf("Done sending entity list")
			} else if strings.HasPrefix(strbytes, "kill ") {
				parts := strings.Split(strbytes, " ")
				log.Printf("Killing entity %v", parts[1])
				killhash, err := strconv.ParseInt(parts[1], 10, 32)
				if err != nil {
					log.Fatal("Could not parse kill hash")
				}
				for _, pod := range podsList.Items {
					entity := formatEntityName(pod)
					if hash(entity) == int32(killhash) {
						log.Printf("Pod to kill: %v", entity)
						clientset.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
						break
					}
				}
				conn.Close()
				stop = true
				log.Printf("Done killing entity ")
			}
		}
	}
}

func main() {
	listener, err := net.Listen("unix", "/dockerdoom.socket")
	if err != nil {
		log.Fatalf("Could not create socket file")
	}

	log.Print("Create virtual display")
	startCmd("/usr/bin/Xvfb :99 -ac -screen 0 640x480x24")
	time.Sleep(time.Duration(2) * time.Second)
	startCmd("x11vnc -geometry 640x480 -forever -usepw -display :99")
	log.Print("You can now connect to it with a VNC viewer at port 5900")

	log.Print("Trying to start DOOM ...")
	startCmd("/usr/bin/env DISPLAY=:99 /usr/local/games/psdoom -skill 1 -nomouse -nosound -file psdoom1.wad -nopslev")

	podsListChan := make(chan *v1.PodList)
	// query the running pods every 5 sec
	go func() {
		clientset := NewClientSet()

		for {
			options := metav1.ListOptions{FieldSelector: "status.phase=Running"}
			podsList, err := clientset.CoreV1().Pods("").List(context.Background(), options)
			podsListChan <- podsList
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	socketLoop(listener, podsListChan)
}
