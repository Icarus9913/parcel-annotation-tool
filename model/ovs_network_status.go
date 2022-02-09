package model

type OvsNetworkStatus struct {
	PoolName       string   `json:"poolName"`
	RuleName       string   `json:"ruleName"`
	VlanID         string   `json:"vlanID"`
	Ipv4           string   `json:"ipv4"`
	Ipv6           string   `json:"ipv6"`
	IsDefaultRoute string   `json:"isDefaultRoute"`
	Interface      string   `json:"interface"`
	DefaultRouteV4 string   `json:"defaultRouteV4"`
	DefaultRouteV6 string   `json:"defaultRouteV6"`
	Route          []string `json:"route"`
}
