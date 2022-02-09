package model

type IPPool struct {
	SubnetName     string   `json:"SubnetName"`
	Subnet         string   `json:"Subnet"`
	DisplayVlan    int      `json:"DisplayVlan"`
	Namespace      string   `json:"Namespace"`
	DefaultRouteV4 string   `json:"DefaultRouteV4"`
	DefaultRouteV6 string   `json:"DefaultRouteV6"`
	Route          []string `json:"Route"`
	PrefixV6       string   `json:"PrefixV6"`
	PrefixMac      string   `json:"PrefixMac"`
	PoolName       string   `json:"PoolName"`
}
