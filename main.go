package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	// "k8s.io/apimachinery/pkg/apis/meta/v1"

	bV1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	return os.Getenv("HOME")
}

func main() {

	jobName := flag.String("j", "", "Job Name to Watch")
	nameSpace := flag.String("n", "default", "Name Space to watch Jobs in")

	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	done := make(chan bool)

	watchlist := cache.NewListWatchFromClient(clientset.BatchV1().RESTClient(), "jobs", *nameSpace,
		fields.OneTermEqualSelector("metadata.name", *jobName))
	_, controller := cache.NewInformer(
		watchlist,
		&bV1.Job{},
		time.Second*60,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				myJob := obj.(*bV1.Job)
				log.Println("Job Added:", myJob.Name)
				if myJob.Status.Active > 0 {
					log.Println(myJob.Name, "Active")
				}
				if myJob.Status.Succeeded > 0 {
					log.Println(myJob.Name, "Succeeded")
					done <- true
				}
				if myJob.Status.Failed > 0 {
					log.Println(myJob.Name, "Failed")
				}

			},
			DeleteFunc: func(obj interface{}) {
				myJob := obj.(*bV1.Job)
				log.Println("Job Deleted:", myJob.Name)
			},
			UpdateFunc: func(_, obj interface{}) {
				myJob := obj.(*bV1.Job)
				log.Println("Job Updated:", myJob.Name)
				if myJob.Status.Active > 0 {
					log.Println(myJob.Name, "Active")
				}
				if myJob.Status.Succeeded > 0 {
					log.Println(myJob.Name, "Succeeded")
					done <- true
				}
				if myJob.Status.Failed > 0 {
					log.Println(myJob.Name, "Failed")
				}
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)
	log.Println("Started Watching", *jobName, "Job")
	<-done
	close(stop)
	log.Println(*jobName, "Completed")

}
