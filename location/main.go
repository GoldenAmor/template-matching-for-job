package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
)

var locationX int
var locationY int
var url string

func init() {
	flag.IntVar(&locationX, "x", 0, "指定坐标的x轴坐标")
	flag.IntVar(&locationY, "y", 0, "指定坐标的y轴坐标")
	flag.StringVar(&url, "method", "concurrent", "指定调用同步/异步接口")
	flag.Parse()
	url = "http://localhost:6060/template-match-" + url
}

type Response struct {
	Data []image.Rectangle `json:"data"`
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
}

func main() {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(data.Data); i++ {
		if data.Data[i].Max.X < locationX || data.Data[i].Max.Y < locationY || data.Data[i].Min.X > locationX || data.Data[i].Min.Y > locationY {
			continue
		} else {
			fmt.Printf("指定坐标属于第%d个元素", i+1)
			return
		}
	}
	fmt.Println("指定坐标不在模板元素范围内！")
	return
}
