package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Service struct {
	Name       string `yaml:"name"`
	RemotePort int    `yaml:"remotePort"`
	LocalPort  int    `yaml:"localPort"`
}
type ServiceList struct {
	Services []Service `yaml:"services"`
}

func main() {

	stopChs := []chan struct{}{}
	readyChs := []chan struct{}{}
	//defer close(stopCh)
	var wg sync.WaitGroup
	wg.Add(1)
	// managing termination signal from the terminal. As you can see the stopCh
	// gets closed to gracefully handle its termination.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Bye...")
		for _, ch := range stopChs {
			close(ch)
		}
		wg.Done()
	}()

	// Get the path to the kubeconfig file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	// Build the clientset from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// read service.yaml
	var serviceList ServiceList
	yamlFile, err := ioutil.ReadFile("services.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, &serviceList)
	if err != nil {
		log.Fatal(err)
	}

	for _, serviceName := range serviceList.Services {

		// Start the port forwarding
		go func(serviceName Service) {
			podName, port := findPod(clientset, serviceName.Name)
			path := fmt.Sprintf("api/v1/namespaces/%s/pods/%s/portforward", "default", podName)
			svcURL, err := url.Parse(fmt.Sprintf("%s/%v", config.Host, path))
			if err != nil {
				log.Fatal("url parse", err)
			}
			fmt.Println(svcURL)
			transport, upgrader, err := spdy.RoundTripperFor(config)
			if err != nil {
				log.Fatal("spdy roundtripper", err)
			}
			stopCh := make(chan struct{}, 1)
			readyCh := make(chan struct{})
			readyChs = append(readyChs, readyCh)
			stopChs = append(stopChs, stopCh)
			dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, svcURL)
			portForwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%v", serviceName.LocalPort, port)}, stopCh, readyCh, os.Stdout, os.Stderr)
			if err != nil {
				log.Fatal("error forwarding", err)
			}
			if err := portForwarder.ForwardPorts(); err != nil {
				log.Fatal("port-forwarder", err)
			}
		}(serviceName)

		fmt.Printf("Port forwarding to service '%s' started.\n", serviceName.Name)
		fmt.Println("You can now access the service on http://localhost:" + fmt.Sprintf("%d", serviceName.LocalPort))
	}
	fmt.Println("Press Ctrl+C to stop forwarding.")
	wg.Wait()
	fmt.Println("Bye...")
}

// findPod finds a pod from a service name
func findPod(clientset *kubernetes.Clientset, serviceName string) (string, string) {
	// Retrieve the service object
	service, err := clientset.CoreV1().Services("default").Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// Get the labels associated with the service
	serviceLabels := service.Spec.Selector

	// Retrieve the pods with matching labels
	pods, err := clientset.CoreV1().Pods(service.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: serviceLabels}),
	})
	if err != nil {
		log.Fatal(err)
	}

	return pods.Items[0].Name, fmt.Sprintf("%v", pods.Items[0].Spec.Containers[0].Ports[0].ContainerPort)

}
