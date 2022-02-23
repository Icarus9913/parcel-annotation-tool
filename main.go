package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"parcel-annotation-tool/model"
	"parcel-annotation-tool/util"
	"strconv"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
	"go.etcd.io/etcd/clientv3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

var log = logging.Logger("main")

var ADD_KEY = "dce.daocloud.io/parcel.ovs.network.status"
var OVS_POOL_ETCD_PATH = "/parcel/ovs_service/pool"
var ETCD_V6V4MAP_KEY = "/parcel/ovs_service/v6v4map"
var PARCEL_NET_TYPE = "dce.daocloud.io/parcel.net.type"
var PARCEL_NET_VALUE = "dce.daocloud.io/parcel.net.value"

var runningMode bool

func init() {
	_ = logging.SetLogLevel("*", "INFO")
	util.Initialize()
}

func main() {
	startCMD()

}

func startCMD() {
	app := &cli.App{
		Name:  "parcel upgrade from 4.0.3 to 4.0.7 annotation adding tool",
		Usage: "",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "really-do-it",
				Usage: "pass the flag if you wanna patch the annotation with tool automatation",
			},
			&cli.StringFlag{
				Name:  "namespace",
				Usage: "pass the specific namespace that you need to patch the annotation",
				Value: "default",
			},
			&cli.StringFlag{
				Name:  "podName",
				Usage: "pass the specific podName that you need to patch the annotation",
				Value: "",
			},
		},
		Commands: []*cli.Command{
			runCmd,
		},
	}
	err := app.Run(os.Args)
	if nil != err {
		log.Fatal(err)
	}
}

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "start running tool",
	Action: func(cctx *cli.Context) error {
		ctx := context.TODO()

		runningMode = cctx.Bool("really-do-it")
		namespace := cctx.String("namespace")
		podName := cctx.String("podName")
		err := run(ctx, namespace, podName)
		if nil != err {
			return err
		}
		return nil
	},
}

func run(ctx context.Context, namespace, podName string) error {
	clientset, err := util.NewK8s()
	if nil != err {
		return errors.New("connect to k8s failed with err: " + err.Error())
	}
	hostname, err := os.Hostname()
	if nil != err {
		return err
	}
	etcdurl := ""

	// 获取nodeName
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if nil != err {
		log.Error(err)
		return err
	}
	for _, n := range nodeList.Items {
		if n.Name == hostname {
			for _, v := range n.Status.Addresses {
				if v.Type == v1.NodeInternalIP {
					etcdurl = v.Address
				}
			}
		}
	}

	endpoints, err := clientset.CoreV1().Endpoints("kube-system").Get(ctx, "dce-etcd", metav1.GetOptions{})
	if nil != err {
		log.Error(err)
		return err
	}
	for _, sub := range endpoints.Subsets {
		for _, v := range sub.Ports {
			if v.Name == "client-https" {
				etcdurl = fmt.Sprint(etcdurl, ":", v.Port)
			}
		}
	}

	util.ETCD_URL = etcdurl

	/*	list, err := clientset.DiscoveryV1beta1().EndpointSlices("kube-system").List(ctx, metav1.ListOptions{})
		if nil != err {
			log.Error(err)
			return err
		}
		for _, v := range list.Items {

			if strings.HasPrefix(v.Name, "dce-etcd") {
				for _, e := range v.Endpoints {
					if *e.Hostname == nodeName {
						etcdurl = e.Addresses[0]
						break
					}
				}

				for _, vv := range v.Ports {
					if *vv.Name == "client-https" {
						etcdurl = fmt.Sprint(etcdurl, *vv.Port)
						break
					}
				}
			}
		}*/

	// specific pod
	if "" != podName {
		log.Warnf("Start to get namespace %s specific pod %s annotation", namespace, podName)
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if nil != err {
			return err
		}
		if _, ok := pod.Annotations[ADD_KEY]; !ok {
			collect_stupid_data(ctx, clientset, *pod)
		}
		return nil
	}

	// pod list
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if nil != err {
		return err
	}

	log.Warnf("Start to get namespace %s pods list annotation", namespace)
	for _, pod := range podList.Items {
		if _, ok := pod.Annotations[ADD_KEY]; !ok {
			collect_stupid_data(ctx, clientset, pod)
		}
	}
	return nil
}

func collect_stupid_data(ctx context.Context, clientset *kubernetes.Clientset, pod v1.Pod) {
	tmp_pod_ns := pod.Namespace
	tmp_pod_ip := pod.Status.PodIP
	tmp_pod_net_type := pod.Annotations[PARCEL_NET_TYPE]
	if tmp_pod_net_type == "ovs" {
		//tmp_annotation := map[string]interface{}{}
		var tmp_annotation model.OvsNetworkStatus

		// get poolName and ruleName
		tmp_pod_net_value := pod.Annotations[PARCEL_NET_VALUE]
		tmp_split := strings.Split(tmp_pod_net_value, ":")

		if strings.HasPrefix(tmp_pod_net_value, "rule") {
			//tmp_annotation["poolName"], tmp_annotation["ruleName"] = tmp_split[1], tmp_split[2]
			tmp_annotation.PoolName, tmp_annotation.RuleName = tmp_split[1], tmp_split[2]
		} else if strings.HasPrefix(tmp_pod_net_value, "pool") {
			//tmp_annotation["poolName"], tmp_annotation["ruleName"] = tmp_split[1], ""
			tmp_annotation.PoolName, tmp_annotation.RuleName = tmp_split[1], ""
		} else {
			log.Errorf("pod %v seems like wrong, please check it!", pod.Name)
			return
		}

		if tmp_annotation.PoolName != "" {
			var poolInfo model.IPPool
			etcdV3Client, err := util.NewEtcd()
			if nil != err {
				log.Errorf("connect to ETCD failed with err %v", err)
				return
			}
			response, err := etcdV3Client.EtcdClient.Get(ctx, OVS_POOL_ETCD_PATH, clientv3.WithPrefix())
			if nil != err {
				log.Errorf("get pod %v ippool information from etcd failed with err  ", pod.Name, err.Error())
				return
			}

			err = json.Unmarshal(response.Kvs[0].Value, &poolInfo)
			if nil != err {
				log.Errorf("Unmarshal pod %v ippool %v with err %v", pod.Name, tmp_annotation.PoolName, err)
				return
			}

			/*			tmpshit, err := json.Marshal(poolInfo)
						if nil != err {
							log.Error("marshal挂了, ", err)
							return
						}
						log.Info("打印看看: ", string(tmpshit))*/

			// vlanID
			tmp_annotation.VlanID = strconv.Itoa(poolInfo.DisplayVlan)

			if "" != poolInfo.Subnet {
				ip_v4_network_prefix := strings.Split(poolInfo.Subnet, "/")[1]
				// ipv4
				tmp_annotation.Ipv4 = strings.Join([]string{tmp_pod_ip, ip_v4_network_prefix}, "/")
			}

			if "" != poolInfo.PrefixV6 {
				ip_v6_network_prefix := strings.Split(poolInfo.PrefixV6, "/")[1]
				tmp_ipv6 := ""
				v6v4Map, err := v6_v4_map(ctx)
				if nil != err {
					log.Error("Getting v6_v4_map failed, ", err)
					return
				}

				for k := range v6v4Map {
					if v6v4Map[k] == tmp_pod_ip {
						tmp_ipv6 = k
						break
					}
				}
				if "" != tmp_ipv6 {
					// ipv6
					tmp_annotation.Ipv6 = strings.Join([]string{tmp_ipv6, ip_v6_network_prefix}, "/")
				}
			}

			// isDefaultRoute
			tmp_annotation.IsDefaultRoute = "true"

			// interface
			tmp_annotation.Interface = "eth0"

			// defaultRouteV4
			tmp_annotation.DefaultRouteV4 = poolInfo.DefaultRouteV4

			// defaultRouteV6
			tmp_annotation.DefaultRouteV6 = poolInfo.DefaultRouteV6

			// route
			if len(poolInfo.Route) == 0 {
				tmp_annotation.Route = []string{""}
			} else {
				tmp_annotation.Route = poolInfo.Route
			}
		}

		marshalData, err := json.Marshal([]interface{}{tmp_annotation})
		if nil != err {
			log.Errorf("Marshal pod %v ippool %v with err %v", pod.Name, tmp_annotation.PoolName, err)
			return
		}
		if runningMode {
			pod.Annotations[ADD_KEY] = string(marshalData)
			var anno = map[string]string{ADD_KEY: "'" + string(marshalData) + "'"}
			patchData, err := json.Marshal(anno)
			if nil != err {
				log.Errorf("Marshal pod %v patchData with err %v", pod.Name, err)
				return
			}

			_, err = clientset.CoreV1().Pods(tmp_pod_ns).Patch(ctx, pod.Name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
			if nil != err {
				log.Errorf("patch pod %v annotation %v with err %v", pod.Name, string(marshalData), err)
				return
			}
		} else {
			anno := model.AnnotationPatch{model.Metadata{model.Annotations{string(marshalData)}}}
			patchResult, err := json.Marshal(anno)
			if nil != err {
				log.Errorf("Marshal pod %v annotation %v with err %v", pod.Name, string(marshalData), err)
				return
			}
			log.Infof("Please run the command by yourself: kubectl patch pod %s --patch '%s'", pod.Name, string(patchResult))
		}
	}
}

// ipv6_ipv4_map
func v6_v4_map(ctx context.Context) (map[string]string, error) {
	etcdV3Client, err := util.NewEtcd()
	if nil != err {
		return nil, errors.New("connect to ETCD failed with err: " + err.Error())
	}
	response, err := etcdV3Client.EtcdClient.Get(ctx, ETCD_V6V4MAP_KEY)
	if nil != err {
		return nil, err
	}

	if len(response.Kvs) == 0 {
		return nil, nil
	}
	var v6v4map = map[string]string{}
	for _, v := range response.Kvs {
		err := json.Unmarshal(v.Value, &v6v4map)
		if nil != err {
			return nil, err
		}
	}
	return v6v4map, nil
}
