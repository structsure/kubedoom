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

	"k8s.io/apimachinery/pkg/api/errors"
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

type Mode interface {
	getEntities() Results
	deleteEntity(string)
}

type podmode struct {
}
type Results []string

func (r Results) Only(matching string) Results {
	return r.filterResults(
		func(inThisString string) bool {
			return strings.Contains(inThisString, matching)
		})
}
func (r Results) Not(matching string) Results {
	return r.filterResults(
		func(inThisString string) bool {
			return !strings.Contains(inThisString, matching)
		})
}
func (r Results) filterResults(filter func(string) bool) Results {
	filtered := []string{}
	for iSlice, vSlice := range r {
		if filter(vSlice) {
			filtered = append(filtered, r[iSlice])
		}
	}
	return filtered
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
func ExampleGetEntitiesK8sClient() []string {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		var listoptions = metav1.ListOptions{
			Limit:    500,
			Continue: "true",
		}
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), listoptions)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods(os.Getenv("NAMESPACE")).Get(context.TODO(), "", metav1.GetOptions{ResourceVersion: ""})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found example-xxxxx pod in default namespace\n")
		}

		time.Sleep(10 * time.Second)
		return []string{"this"}
	}
}

func dontPanic[a any](ret *a, err error) *a {
	if err != nil {
		panic(err.Error())
	}
	return ret
}
func getEntitiesK8sClient() []string {

	clientset := kubernetes.NewForConfigOrDie(dontPanic(rest.InClusterConfig()))
	var listoptions = metav1.ListOptions{
		// LabelSelector: "",
	}
	pods := dontPanic(clientset.CoreV1().Pods("").List(context.TODO(), listoptions))
	// pods := dontPanicPTR(clientset.CoreV1().Pods("").List(context.TODO(), listoptions))
	log.Printf("PODLIST looks like this")
	log.Printf("%v", pods)
	return []string{"this"}
}

func (m podmode) getEntities() Results {
	var args []string
	if namespace, exists := os.LookupEnv("NAMESPACE"); exists {
		args = []string{"kubectl", "get", "pods", "--namespace", namespace, "-o", "go-template",
			"--template={{range .items}}{{.metadata.namespace}}/{{.metadata.name}} {{end}}"}
	} else {
		args = []string{"kubectl", "get", "pods", "-A", "-o", "go-template", "--template={{range .items}}{{.metadata.namespace}}/{{.metadata.name}} {{end}}"}
	}
	getEntitiesK8sClient()
	output := outputCmd(args)
	outputstr := strings.TrimSpace(output)
	pods := strings.Split(outputstr, " ")
	log.Printf("Pods in Get Entities: %v", pods)
	return pods
}
func NsAndPod(entity string) (string, string) {
	podparts := strings.Split(entity, "/")
	return podparts[0], podparts[1]
}
func LabelPod(applyLabel, ns, pod string) (string, string) {
	log.Printf("Applying label: %v to %v/%v", applyLabel, ns, pod)
	cmd := exec.Command("/usr/bin/kubectl", "label", "pods", "-n", ns, pod, applyLabel)
	go cmd.Run()
	return ns, pod
}
func (m podmode) deletePod(ns, pod string) {
	log.Printf("Pod %v in Namespace %v to kill", pod, ns)
	cmd := exec.Command("/usr/bin/kubectl", "delete", "pod", "-n", ns, pod)
	go cmd.Run()
}
func (m podmode) deleteEntity(entity string) {
	log.Printf("Entity to kill: %v", entity)
	ns, pod := NsAndPod(entity)
	LabelPod("KilledBy="+TryEnv("Player"), ns, pod)
	cmd := exec.Command("/usr/bin/kubectl", "delete", "pod", "-n", ns, pod)
	go cmd.Run()
}

type nsmode struct {
}

func (m nsmode) getEntities() Results {
	args := []string{"kubectl", "get", "namespaces", "-o", "go-template", "--template={{range .items}}{{.metadata.name}} {{end}}"}
	output := outputCmd(args)
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
			entities := mode.getEntities().Only("kubedoom").Not(Me())
			if strbytes == "list" {
				for _, entity := range entities {
					padding := strings.Repeat("\n", 255-len(entity))
					_, err = conn.Write([]byte(entity + padding))
					if err != nil {
						log.Fatal("Could not write to socker file")
					}
				}
			} else if strings.HasPrefix(strbytes, "kill ") {
				parts := strings.Split(strbytes, " ")
				killhash, err := strconv.ParseInt(parts[1], 10, 32)
				if err != nil {
					log.Fatal("Could not parse kill hash")
				}
				for _, entity := range entities {
					if hash(entity) == int32(killhash) {
						log.Printf("calling delete entry for %v", entity)
						mode.deleteEntity(entity)
						break
					}
				}
			} else {
				log.Printf("received %v from strbytes", strbytes)
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
