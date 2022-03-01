package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/cert"
)

var log = logging.Logger("util")

var (
	ScInPodPath    = "/var/run/secrets/kubernetes.io/serviceaccount"
	KubeConfigPath = filepath.Join(os.Getenv("HOME"), ".kube", "config")
)

func ExistFile(filePath string) bool {
	if info, err := os.Stat(filePath); err == nil {
		if !info.IsDir() {
			return true
		}
	}
	return false
}

func ExistDir(dirPath string) bool {
	if info, err := os.Stat(dirPath); err == nil {
		if info.IsDir() {
			return true
		}
	}
	return false
}

func autoConfig() (*rest.Config, error) {

	var config *rest.Config
	var err error

	if ExistFile(KubeConfigPath) == true {

		config, err = clientcmd.BuildConfigFromFlags("", KubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get config from kube config=%v , info=%v", KubeConfigPath, err)
		}

	} else if ExistDir(ScInPodPath) == true {

		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get config from serviceaccount=%v , info=%v", ScInPodPath, err)
		}

	} else {
		return nil, fmt.Errorf("failed to get config ")
	}

	return config, nil
}

func NewK8s() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	config, err = autoConfig()
	//config, err = InClusterConfig()

	if nil != err {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func InClusterConfig() (*rest.Config, error) {
	token, err := ioutil.ReadFile(KUBE_TOKEN)
	if err != nil {
		return nil, err
	}

	tlsClientConfig := rest.TLSClientConfig{}

	if _, err := cert.NewPool(KUBE_CA); err != nil {
		log.Errorf("Expected to load root CA config from %s, but got err: %v", KUBE_CA, err)
	} else {
		tlsClientConfig.CAFile = KUBE_CA
	}

	return &rest.Config{
		// TODO: switch to using cluster DNS.
		Host:            KUBE_URL,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(token),
		BearerTokenFile: KUBE_TOKEN,
	}, nil
}
