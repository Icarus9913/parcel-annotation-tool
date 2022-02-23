package util

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
)

func TestSTH(t *testing.T) {
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}

	fmt.Println(*kubeconfig)
}

func TestFake(t *testing.T) {
	ctx := context.TODO()
	clientset := fake.NewSimpleClientset()
	podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if nil != err {
		panic(err)
	}
	fmt.Println(podList)
}
