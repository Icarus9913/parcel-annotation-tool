package main

import (
	"encoding/json"
	"fmt"
)

type Dog struct {
	Result string `json:"result"`
}

type Man struct {
	Name string `json:"name"`
}

func main() {
	d := Dog{Result: "wangwang"}
	dd, _ := json.Marshal(d)

	h := Man{Name: string(dd)}
	data, err := json.Marshal(h)
	if nil != err {
		panic(err)
	}
	fmt.Println(string(data))
}
