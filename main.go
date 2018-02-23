package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"strings"
	"sync"
	"time"

	mvutil "podmove/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

//global variables
var (
	masterUrl            string
	kubeConfig           string
	nameSpace            string
	podName              string
	nodeName             string
	k8sVersion           string
)

const (
	defaultRetryLess              = 2
	defaultSleep                  = time.Second * 10
	defaultWaitLockTimeOut        = time.Second * 100
)

func setFlags() {
	flag.StringVar(&masterUrl, "masterUrl", "", "master url")
	flag.StringVar(&kubeConfig, "kubeConfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&nameSpace, "nameSpace", "default", "kubernetes object namespace")
	flag.StringVar(&podName, "podName", "myschedule-cpu-80", "the podNames to be handled, split by ','")
	flag.StringVar(&nodeName, "nodeName", "", "Destination of move")
	flag.StringVar(&k8sVersion, "k8sVersion", "1.6", "the version of Kubenetes cluster, candidates are 1.5 | 1.6")

	flag.Set("alsologtostderr", "true")
	flag.Parse()
}

func movePod(client *kclient.Clientset, nameSpace, podName, nodeName string) (*v1.Pod, error) {
	podClient := client.CoreV1().Pods(nameSpace)
	id := fmt.Sprintf("%v/%v", nameSpace, podName)

	//1. get original Pod
	getOption := metav1.GetOptions{}
	pod, err := podClient.Get(podName, getOption)
	if err != nil {
		err = fmt.Errorf("move-aborted: get original pod:%v\n%v", id, err.Error())
		glog.Error(err.Error())
		return nil, err
	}

	if pod.Spec.NodeName == nodeName {
		err = fmt.Errorf("move-aborted: pod %v is already on node: %v", id, nodeName)
		glog.Error(err.Error())
		return nil, err
	}

	glog.V(2).Infof("move-pod: begin to move %v from %v to %v",
		id, pod.Spec.NodeName, nodeName)

	//2. invalidate the schedulerName of parent controller
	parentKind, parentName, err := mvutil.ParseParentInfo(pod)
	if err != nil {
		return nil, fmt.Errorf("move-abort: cannot get pod-%v parent info: %v", id, err.Error())
	}

	//2.1 if pod is barely standalone pod, move it directly
	if parentKind == "" {
		return mvutil.MoveBarePod(client, pod, nodeName, defaultRetryLess)
	}

	//2.2 if pod controlled by ReplicationController/ReplicaSet, then need to do more
	return mvutil.MovePod(client, pod, parentKind, parentName, nodeName)
}

func movePods(client *kclient.Clientset, nameSpace, podNames, nodeName string) error {
	names := strings.Split(podNames, ",")
	var wg sync.WaitGroup

	for _, pname := range names {
		podName := strings.TrimSpace(pname)
		if len(podName) == 0 {
			continue
		}
		wg.Add(1)

		go func() {
			defer wg.Done()
			rpod, err := movePod(client, nameSpace, podName, nodeName);
			if err != nil {
				glog.Errorf("move pod[%s] failed: %v", podName, err)
				return
			}

			glog.V(2).Infof("sleep 10 seconds to check the final state")
			time.Sleep(time.Second * 10)
			if err := mvutil.CheckPodMoveHealth(client, nameSpace, rpod.Name, nodeName); err != nil {
				glog.Errorf("move pod[%s] failed: %v", podName, err)
				return
			}
			glog.V(2).Infof("move pod(%v/%v) to node-%v successfully", nameSpace, podName, nodeName)
		}()
	}

	wg.Wait()
	return nil
}

func main() {
	setFlags()
	defer glog.Flush()

	kubeClient := mvutil.GetKubeClient(masterUrl, kubeConfig)
	if kubeClient == nil {
		glog.Errorf("failed to get a k8s client for masterUrl=[%v], kubeConfig=[%v]", masterUrl, kubeConfig)
		return
	}

	if nodeName == "" {
		glog.Errorf("nodeName should not be empty.")
		return
	}

	if err := movePods(kubeClient, nameSpace, podName, nodeName); err != nil {
		glog.Errorf("move pod failed: %v/%v, %v", nameSpace, podName, err.Error())
		return
	}

}
