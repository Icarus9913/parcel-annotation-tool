package util

import (
	"strings"
	"time"

	"github.com/coreos/etcd/pkg/transport"
	"go.etcd.io/etcd/clientv3"
)

type EtcdV3Client struct {
	EtcdClient *clientv3.Client
}

func NewEtcd() (EtcdV3Client, error) {
	log.Info("ETCD_CERT: ", ETCD_CERT)
	log.Info("ETCD_KEY: ", ETCD_KEY)
	log.Info("ETCD_CA: ", ETCD_CA)
	log.Info("ETCD_URL: ", ETCD_URL)

	tlsInfo := transport.TLSInfo{
		CertFile:      ETCD_CERT,
		KeyFile:       ETCD_KEY,
		TrustedCAFile: ETCD_CA,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if nil != err {
		return EtcdV3Client{}, err
	}

	cfg := clientv3.Config{
		Endpoints:   strings.Split(ETCD_URL, ","),
		TLS:         tlsConfig,
		DialTimeout: time.Second * 5,
	}
	client, err := clientv3.New(cfg)
	if nil != err {
		return EtcdV3Client{}, err
	}
	return EtcdV3Client{EtcdClient: client}, nil
}
