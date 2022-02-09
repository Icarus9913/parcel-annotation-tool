package model

type AnnotationPatch struct {
	Metadata `json:"metadata"`
}

type Metadata struct {
	Annotations `json:"annotations"`
}

type Annotations struct {
	Key string `json:"dce.daocloud.io/parcel.ovs.network.status"`
}
