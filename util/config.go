package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// etcd-configuration
var ETCD_URL = "https://1.1.1.1:12379"
var ETCD_CA = "./certs/etcd-secrets/etcd-ca"
var ETCD_KEY = "./certs/etcd-secrets/etcd-key"
var ETCD_CERT = "./certs/etcd-secrets/etcd-cert"

var ETCD_CER_Prefix = "etcd-secrets"

// kubernetes-configuration
var KUBE_URL = "https://1.1.1.1:11081"
var KUBE_CA = "./certs/k8s-secrets/ca.crt"
var KUBE_TOKEN = "./certs/k8s-secrets/token"
var KUBE_NAMESPACE = "./certs/k8s-secrets/namespace"

var (
	EtcdConfigCaPath   string
	EtcdConfigCertPath string
	EtcdConfigKeyPath  string
	EtcdConfigEndpoint string
)

var envToConfig = map[string]*string{
	"ETCDCTL_CACERT":   &EtcdConfigCaPath,
	"ETCDCTL_CERT":     &EtcdConfigCertPath,
	"ETCDCTL_KEY":      &EtcdConfigKeyPath,
	"PARCEL_ETCD_LIST": &EtcdConfigEndpoint,
}

func initialize() {
	for name, point := range envToConfig {
		val := os.Getenv(name)

		if len(val) == 0 {
			log.Warnf("ENV %s has no value", name)
		} else {
			*point = val
		}
	}

}

func Initialize() {
	//configMapPath := os.Getenv("ENV_SYSTEM__CONFIGMAP_PATH")

	// 获取ETCD URL
	etcd_host, etcd_https_port := os.Getenv("DCE_ETCD_SERVICE_HOST"), os.Getenv("DCE_ETCD_SERVICE_PORT_CLIENT_HTTPS")
	if etcd_host != "" && etcd_https_port != "" {
		ETCD_URL = fmt.Sprint("https://", etcd_host, ":", etcd_https_port)
		log.Info("Get ETCD URL from env: ", ETCD_URL)
	} else if etcdUrl, err := GetEtcdUrl(); nil == err {
		ETCD_URL = *etcdUrl
		log.Info("Get ETCD URL from K8s Endpoints: ", ETCD_URL)
	} else {
		log.Info("Use the default ETCD URL: ", ETCD_URL)
	}

	if !IsDir(ETCD_CER_Prefix) {
		log.Info("不存在/etcd-secrets证书目录文件, 使用默认配置")
		return
	}

	log.Infof("本地存在%s证书目录文件", ETCD_CER_Prefix)

	ETCD_CA, ETCD_CERT, ETCD_KEY = "", "", ""
	EtcdConfigCaPath = ETCD_CER_Prefix + "/etcd-ca"
	EtcdConfigCertPath = ETCD_CER_Prefix + "/etcd-cert"
	EtcdConfigKeyPath = ETCD_CER_Prefix + "/etcd-key"

	if existFile(EtcdConfigCaPath) {
		ETCD_CA = EtcdConfigCaPath
	} else {
		log.Error("EtcdConfigCaPath不存在")
		return
	}

	if existFile(EtcdConfigCertPath) {
		ETCD_CERT = EtcdConfigCertPath
	} else {
		log.Error("EtcdConfigCertPath不存在")
		return
	}

	if existFile(EtcdConfigKeyPath) {
		ETCD_KEY = EtcdConfigKeyPath
	} else {
		log.Error("EtcdConfigKeyPath不存在")
		return
	}

}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if nil != err {
		return false
	}
	return s.IsDir()
}

// 从k8s endpoints里找etcd url
func GetEtcdUrl() (*string, error) {
	ctx := context.TODO()

	clientset, err := NewK8s()
	if nil != err {
		return nil, errors.New("connect to k8s failed with err: " + err.Error())
	}
	hostname, err := os.Hostname()
	if nil != err {
		return nil, errors.New("get Hostname failed with err: " + err.Error())
	}

	// 获取etcd url
	etcdurl := "https://"
	tmpHost := ""
	tmpPort := ""
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if nil != err {
		return nil, errors.New("get k8s nodes failed with err: " + err.Error())
	}
	for _, n := range nodeList.Items {
		if n.Name == hostname {
			for _, v := range n.Status.Addresses {
				if v.Type == v1.NodeInternalIP {
					tmpHost = v.Address
					break
				}
			}
		}
	}

	endpoints, err := clientset.CoreV1().Endpoints("kube-system").Get(ctx, "dce-etcd", metav1.GetOptions{})
	if nil != err {
		return nil, errors.New("get k8s endpoints failed with err: " + err.Error())
	}
	for _, sub := range endpoints.Subsets {
		for _, v := range sub.Ports {
			if v.Name == "client-https" {
				tmpPort = strconv.Itoa(int(v.Port))
				break
			}
		}
	}
	if tmpHost != "" && tmpPort != "" {
		etcdurl = fmt.Sprint(etcdurl, tmpHost, ":", tmpPort)
		return &etcdurl, nil
	}
	return nil, errors.New("get etcd url from k8s endpoint failed")
}
