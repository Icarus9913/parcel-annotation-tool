package main

import (
	"bufio"
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
var OVS_POOL_ETCD_PATH = "/parcel/ovs_service/pool/"
var ETCD_V6V4MAP_KEY = "/parcel/ovs_service/v6v4map"
var PARCEL_NET_TYPE = "dce.daocloud.io/parcel.net.type"
var PARCEL_NET_VALUE = "dce.daocloud.io/parcel.net.value"

var shellFile = "anno.sh"
var buf *bufio.Writer

func init() {
	_ = logging.SetLogLevel("*", "INFO")
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
			&cli.BoolFlag{
				Name:  "out-to-shell",
				Usage: "use this flag to output into a shell file named: anno.sh",
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

		runningMode := cctx.Bool("really-do-it")
		namespace := cctx.String("namespace")
		podName := cctx.String("podName")
		out2shell := cctx.Bool("out-to-shell")
		err := run(ctx, namespace, podName, runningMode, out2shell)
		if nil != err {
			return err
		}
		return nil
	},
}

func run(ctx context.Context, namespace, podName string, runningMode, out2shell bool) error {
	util.Initialize()

	clientset, err := util.NewK8s()
	if nil != err {
		return errors.New("connect to k8s failed with err: " + err.Error())
	}

	if out2shell {
		file, err := os.OpenFile(shellFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if nil != err {
			return errors.New("Open file " + shellFile + "failed with err: " + err.Error())
		}
		defer file.Close()
		buf = bufio.NewWriter(file)
	}

	// specific pod
	if podName != "" {
		log.Infof("Start to get namespace %s specific pod %s annotation", namespace, podName)
		pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if nil != err {
			return errors.New(fmt.Sprintf("Getting pod %s from k8s failed with err %v", podName, err))
		}

		netType := pod.Annotations[PARCEL_NET_TYPE]
		if netType != "ovs" {
			log.Warnf("Pod %s net type is %s, skip it!", podName, netType)
			return nil
		}

		if _, ok := pod.Annotations[ADD_KEY]; !ok {
			collect_stupid_data(ctx, clientset, *pod, runningMode, out2shell)
			return nil
		}
		log.Warnf("Pod %s already has the annotation %s, skip it!", podName, ADD_KEY)
		return nil
	}

	// pod list
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if nil != err {
		return err
	}

	log.Warnf("Start to get namespace %s pods list annotation", namespace)
	for _, pod := range podList.Items {
		if pod.Annotations[PARCEL_NET_TYPE] != "ovs" {
			continue
		}

		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		if _, ok := pod.Annotations[ADD_KEY]; !ok {
			collect_stupid_data(ctx, clientset, pod, runningMode, out2shell)
		}
	}
	return nil
}

func collect_stupid_data(ctx context.Context, clientset *kubernetes.Clientset, pod v1.Pod, runningMode, out2shell bool) {
	tmp_pod_ns := pod.Namespace
	tmp_pod_ip := pod.Status.PodIP
	var tmp_annotation model.OvsNetworkStatus

	// get poolName and ruleName
	tmp_pod_net_value := pod.Annotations[PARCEL_NET_VALUE]
	tmp_split := strings.Split(tmp_pod_net_value, ":")

	if strings.HasPrefix(tmp_pod_net_value, "rule") {
		tmp_annotation.PoolName, tmp_annotation.RuleName = tmp_split[1], tmp_split[2]
	} else if strings.HasPrefix(tmp_pod_net_value, "pool") {
		tmp_annotation.PoolName, tmp_annotation.RuleName = tmp_split[1], ""
	} else {
		log.Errorf("pod %v seems like wrong, please check it!", pod.Name)
		return
	}

	if tmp_annotation.PoolName != "" {
		tmp_split_poolName := strings.Split(tmp_annotation.PoolName, "-")
		tmp_subnet_name := tmp_split_poolName[0]
		tmp_ns := tmp_split_poolName[1]

		var poolInfo model.IPPool
		etcdV3Client, err := util.NewEtcd()
		if nil != err {
			log.Errorf("connect to ETCD failed with err %v", err)
			return
		}

		tmp_prefix := fmt.Sprint(OVS_POOL_ETCD_PATH, tmp_subnet_name, "/")
		if tmp_ns != "" {
			tmp_prefix = fmt.Sprint(tmp_prefix, tmp_ns, "/")
		}
		tmp_prefix = fmt.Sprint(tmp_prefix, tmp_annotation.PoolName)
		response, err := etcdV3Client.EtcdClient.Get(ctx, tmp_prefix, clientv3.WithPrefix())
		if nil != err {
			log.Errorf("get pod %s ippool information from etcd failed with err  %v", pod.Name, err.Error())
			return
		}
		if len(response.Kvs) == 0 {
			log.Errorf("Gtting pod %s with no ippool %s data from ETCD, please check ETCD with key: %s", pod.Name, tmp_annotation.PoolName, tmp_prefix)
			return
		}

		err = json.Unmarshal(response.Kvs[0].Value, &poolInfo)
		if nil != err {
			log.Errorf("Unmarshal pod %s ippool %s with err %v", pod.Name, tmp_annotation.PoolName, err)
			return
		}

		// vlanID
		tmp_annotation.VlanID = strconv.Itoa(poolInfo.DisplayVlan)

		if poolInfo.Subnet != "" {
			ip_v4_network_prefix := strings.Split(poolInfo.Subnet, "/")[1]
			// ipv4
			tmp_annotation.Ipv4 = strings.Join([]string{tmp_pod_ip, ip_v4_network_prefix}, "/")
		}

		if poolInfo.PrefixV6 != "" {
			ip_v6_network_prefix := strings.Split(poolInfo.PrefixV6, "/")[1]
			tmp_ipv6 := ""
			v6v4Map, err := v6_v4_map(ctx)
			if nil != err {
				log.Error(fmt.Sprintf("Pod %s Getting v6_v4_map failed %v ", pod.Name, err))
				return
			}

			for k := range v6v4Map {
				if v6v4Map[k] == tmp_pod_ip {
					tmp_ipv6 = k
					break
				}
			}
			if tmp_ipv6 != "" {
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
		log.Errorf("Marshal pod %s ippool % with err %v", pod.Name, tmp_annotation.PoolName, err)
		return
	}

	// 自动帮pod打上annotation
	if runningMode {
		pod.Annotations[ADD_KEY] = string(marshalData)
		var anno = map[string]string{ADD_KEY: "'" + string(marshalData) + "'"}
		patchData, err := json.Marshal(anno)
		if nil != err {
			log.Errorf("Marshal pod %s patchData with err %v", pod.Name, err)
			return
		}

		_, err = clientset.CoreV1().Pods(tmp_pod_ns).Patch(ctx, pod.Name, types.StrategicMergePatchType, patchData, metav1.PatchOptions{})
		if nil != err {
			log.Errorf("patch pod %s annotation %v with err %v", pod.Name, string(marshalData), err)
			return
		}
		return
	}

	anno := model.AnnotationPatch{model.Metadata{model.Annotations{string(marshalData)}}}
	patchResult, err := json.Marshal(anno)
	if nil != err {
		log.Errorf("Marshal pod %s annotation %s with err %v", pod.Name, string(marshalData), err)
		return
	}

	if out2shell {
		tmp_data := fmt.Sprintf("kubectl patch pod %s --patch '%s' \n", pod.Name, string(patchResult))
		_, err := buf.Write([]byte(tmp_data))
		if nil != err {
			log.Errorf("Write [  %s  ] to file with err %v", tmp_data, err)
		}
		buf.Flush()
		return
	}
	log.Infof("Please run the command by yourself: kubectl patch pod %s --patch '%s'", pod.Name, string(patchResult))

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
			return nil, errors.New("Unmarshal ipv6_ipv4_map failed with error: " + err.Error())
		}
	}
	return v6v4map, nil
}
