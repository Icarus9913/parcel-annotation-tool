package util

import (
	"os"
)

// etcd-configuration
var ETCD_URL = "https://1.1.1.1:12379"
var ETCD_CA = "./certs/etcd-secrets/etcd-ca"
var ETCD_KEY = "./certs/etcd-secrets/etcd-key"
var ETCD_CERT = "./certs/etcd-secrets/etcd-cert"

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
	prefix := "etcd-secrets"
	if !IsDir(prefix) {
		log.Info("不存在/etcd-secrets证书目录文件")
	} else {
		log.Info("本地存在etcd-secrets证书目录文件")
	}
	EtcdConfigCaPath = prefix + "/etcd-ca"
	EtcdConfigCertPath = prefix + "/etcd-cert"
	EtcdConfigKeyPath = prefix + "/etcd-key"
	EtcdConfigEndpoint = "https://" + os.Getenv("DCE_ETCD_SERVICE_HOST") + ":" + os.Getenv("DCE_ETCD_SERVICE_PORT_CLIENT_HTTPS")

	if existFile(EtcdConfigCaPath) {
		ETCD_CA = EtcdConfigCaPath
	} else {
		log.Info("EtcdConfigCaPath不存在")
	}

	if existFile(EtcdConfigCertPath) {
		ETCD_CERT = EtcdConfigCertPath
	} else {
		log.Info("EtcdConfigCertPath不存在")
	}

	if existFile(EtcdConfigKeyPath) {
		ETCD_KEY = EtcdConfigKeyPath
	} else {
		log.Infof("EtcdConfigKeyPath不存在")
	}

	if ETCD_URL == "" {
		ETCD_URL = EtcdConfigEndpoint
	}
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if nil != err {
		return false
	}
	return s.IsDir()
}
