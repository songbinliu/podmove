package util

import(
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"

	"github.com/golang/glog"
	"fmt"
	"time"
)

func MoveBarePod(client *kclient.Clientset,  pod *api.Pod, nodeName string, retryNum int) (*api.Pod, error) {
	podName := pod.Namespace + "/" + pod.Name
	glog.V(2).Infof("Begin to move pod: %s", podName)

	//1. clone the original Pod with everything, except a new name;
	npod := &api.Pod{}
	CopyPodInfo(pod, npod)
	npod.Spec.NodeName = nodeName
	npod.Name = GenNewPodName(pod.Name)
	podClient := client.CoreV1().Pods(pod.Namespace)
	rpod, err := podClient.Create(npod)
	if err != nil {
		glog.Errorf("Failed to create a new pod: %s/%s, %v", npod.Namespace, npod.Name, err)
		return nil, err
	}

	//2. delete the original Pod
	delOption := &metav1.DeleteOptions{}
	if err := podClient.Delete(pod.Name, delOption); err != nil {
		glog.Warningf("Move Pod warning: failed to delete original pod(%v): %v", podName, err)
	}

	glog.V(2).Infof("move barePod: %s finished.", podName)
	return rpod, nil
}

func clonePod(client *kclient.Clientset, pod *api.Pod, nodeName string) (*api.Pod, error) {
	npod := &api.Pod{}
	CopyPodWithoutLabel(pod, npod)
	npod.Spec.NodeName = nodeName
	npod.Name = GenNewPodName(pod.Name)

	podClient := client.CoreV1().Pods(pod.Namespace)
	rpod, err := podClient.Create(npod)
	if err != nil {
		glog.Errorf("Failed to create a new pod: %s/%s, %v", npod.Namespace, npod.Name, err)
		return nil, err
	}

	glog.V(2).Infof("Create a new pod success: %s/%s", npod.Namespace, npod.Name)
	glog.V(3).Infof("New pod info: %++v", rpod)

	return rpod, nil
}

//1. create a cloned pod without labels;
//2. kill the original pod;
//3. add labels to the cloned pod;
func MovePod(client *kclient.Clientset, pod *api.Pod, parentKind, parentName, nodeName string) (*api.Pod, error) {
	podName := pod.Namespace + "/" + pod.Name
	glog.V(2).Infof("Begin to move pod(%v), parentKind=%v, parentName=%v", podName, parentKind, parentName)
	podClient := client.CoreV1().Pods(pod.Namespace)
	if podClient == nil {
		err := fmt.Errorf("cannot get Pod client for nameSpace:%v", pod.Namespace)
		glog.Error(err)
		return nil, err
	}

	labels := pod.Labels

	//1. create a cloned pod
	npod, err := clonePod(client, pod, nodeName)
	if err != nil {
		glog.Errorf("Move Pod failed: %v", err)
		return nil, err
	}

	time.Sleep(time.Second * 10)

	//2. delete the original pod
	delOption := &metav1.DeleteOptions{}
	if err := podClient.Delete(pod.Name, delOption); err != nil {
		glog.Warningf("Move Pod warning: failed to delete original pod: %v", err)
	}

	//3. add labels to the cloned pod
	xpod, err := podClient.Get(npod.Name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Move pod failed: failed to get the cloned pod: %v", err)
		return nil, err
	}

	xpod.Labels = labels
	if _, err := podClient.Update(xpod); err != nil {
		glog.Errorf("Move Pod failed: failed to update pod %v", err)
		return nil, err
	}

	glog.Errorf("move Pod finished: %s/%s", xpod.Namespace, xpod.Name)

	return xpod, nil
}
