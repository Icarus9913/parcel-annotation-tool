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
	data := map[string]string{"ni": ""}
	if _, ok := data["ni"]; ok {
		fmt.Println("good")
	} else {
		fmt.Println("no")
	}

}

func jj() {
	d := Dog{Result: "wangwang"}
	dd, _ := json.Marshal(d)

	h := Man{Name: string(dd)}
	data, err := json.Marshal(h)
	if nil != err {
		panic(err)
	}
	fmt.Println(string(data))
}
