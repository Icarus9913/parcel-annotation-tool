package util

import (
	"flag"
	"fmt"
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
