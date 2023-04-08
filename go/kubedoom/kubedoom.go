package main

import (
	"context"
	"flag"
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
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
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

type Mode interface {
	getEntities(chan Entity)
	deleteEntity(Entity)
}

type podmode struct {
}

func RemoveIfFiltered(slice []string, allFilters []func(string) bool) []string {
	filtered := []string{}
	for iSlice, vSlice := range slice {
		for _, vFilter := range allFilters {
			if vFilter(vSlice) {
				filtered = append(filtered, slice[iSlice])
			}
		}
	}
	return filtered
}
func TryEnv(somVar string) string {
	envVal, exists := os.LookupEnv(somVar)
	if !exists {
		log.Printf("%v is not Set", somVar)
	}
	return envVal
}
func Me() string {
	return TryEnv("HOSTNAME")
}
func RemoveIfPresent(slice []string, check string) []string {
	removed := []string{}
	for i, v := range slice {
		if !strings.Contains(v, check) {
			removed = append(removed, slice[i])
		}
	}
	return removed
}

func dontPanic[a any](ret *a, err error) *a {
	if err != nil {
		panic(err.Error())
	}
	return ret
}
func GetClientSet() *kubernetes.Clientset {
	return kubernetes.NewForConfigOrDie(dontPanic(rest.InClusterConfig()))
}
func ListPodsWithLabel(labels string) *v1.PodList {
	return dontPanic(GetClientSet().CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{LabelSelector: labels}))
}
func (m podmode) getEntities(e chan Entity) {
	for _, pod := range ListPodsWithLabel("").Items {
		e <- Entity{pod.Namespace, pod.Name, pod.Status.Message}
	}
	close(e)
}
func (e Entity) toPS() string {
	return fmt.Sprintf("%v/%v", e.ns, e.pod)
}
func (e Entity) toNsAndPod() (string, string) {
	return e.ns, e.pod
}

//	func RunningPods(pod v1.PodInterface) {
//		// pod.
//	}
func LabelPod(ns, pod string) (string, string) {
	log.Printf("Applying label to %v/%v", ns, pod)
	vpod := dontPanic(GetClientSet().CoreV1().Pods(ns).Get(context.TODO(), pod, metav1.GetOptions{}))
	log.Printf("Pod %v", vpod)
	podConfig := dontPanic(
		corev1.ExtractPod(vpod, "KILLER"))
	addme := make(map[string]string)
	addme["KilledBy"] = TryEnv("Player")
	podConfig.WithLabels(addme)
	dontPanic(GetClientSet().CoreV1().Pods(ns).Apply(context.TODO(), podConfig, metav1.ApplyOptions{FieldManager: "KILLER"}))
	return ns, pod
}

func DeletePod(ns, pod string) {
	GetClientSet().CoreV1().Pods(ns).Delete(context.TODO(), pod, metav1.DeleteOptions{})
}
func (m podmode) deleteEntity(entity Entity) {
	log.Printf("Entity to kill: %v", entity.toPS())
	ns, pod := entity.toNsAndPod()
	// LabelPod(ns, pod)
	DeletePod(ns, pod)
}

type nsmode struct {
}

func (m nsmode) getEntities(c chan Entity) {
	args := []string{"kubectl", "get", "namespaces", "-o", "go-template", "--template={{range .items}}{{.metadata.name}} {{end}}"}
	output := outputCmd(args)
	outputstr := strings.TrimSpace(output)
	for _, namespace := range strings.Split(outputstr, " ") {
		c <- Entity{ns: namespace}
	}
	close(c)
}

func (m nsmode) deleteEntity(entity Entity) {
	log.Printf("Namespace to kill: %v", entity)
	exec.Command("/usr/bin/kubectl", "delete", "namespace", entity.ns).Run()
}

type Entity struct {
	ns     string
	pod    string
	status string
}

func (entity Entity) Only(matching string) Entity {
	if strings.Contains(entity.pod, matching) || strings.Contains(entity.ns, matching) {
		return entity
	}
	return Entity{}
}
func (entity Entity) Not(matching string) Entity {
	if !strings.Contains(entity.pod, matching) && !strings.Contains(entity.ns, matching) {
		return entity
	}
	return Entity{}
}
func socketLoop(listener net.Listener, mode Mode) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicf("calling panic with error: %v", err)
		}
		stop := false
		for !stop {
			bytes := make([]byte, 40960)
			n, err := conn.Read(bytes)
			if err != nil {
				log.Printf("error reading bytes %v", n)
				stop = true
			}
			bytes = bytes[0:n]
			strbytes := strings.TrimSpace(string(bytes))
			entityChannel := make(chan Entity)
			go mode.getEntities(entityChannel)
			for entity := range entityChannel {
				log.Printf("entity: %v", entity)
				// if entity.Only("kubedoom").Not(Me()).toPS() == "/" {
				if entity.Not(Me()).toPS() == "/" {
					continue
				}
				// log.Printf("Found an entity: %v", entity.toPS())
				entityString := entity.toPS()
				if strbytes == "list" {
					padding := strings.Repeat("\n", 255-len(entityString))
					go conn.Write([]byte(entityString + padding))
				} else if strings.HasPrefix(strbytes, "kill ") {
					parts := strings.Split(strbytes, " ")
					killhash, err := strconv.ParseInt(parts[1], 10, 32)
					if err != nil {
						log.Fatal("Could not parse kill hash")
					}
					if hash(entityString) == int32(killhash) {
						log.Printf("calling delete entry for %v", entity)
						go mode.deleteEntity(entity)
						break
					}
				} else {
					log.Printf("received %v from strbytes", strbytes)
				}
			}
			conn.Close()
			stop = true
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
	startCmd("/usr/bin/Xvfb :99 -ac -screen 0 640x480x24")
	time.Sleep(time.Duration(2) * time.Second)
	startCmd("x11vnc -geometry 640x480 -verbose -forever -usepw -display :99")
	log.Print("You can now connect to it with a VNC viewer at port 5900")

	log.Print("Trying to start DOOM ...")
	startCmd("/usr/bin/env DISPLAY=:99 /usr/local/games/psdoom -warp -E1M1 -skill 1 -nomouse")
	socketLoop(listener, mode)
}
