package main

import (
	"encoding/json"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Data []image.Rectangle `json:"data"`
	Code string            `json:"code"`
	Msg  string            `json:"msg"`
}

func main() {
	resp, err := http.Get("http://localhost:6060/template-match")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	srcImage := gocv.IMRead("./images/src1.jpg", gocv.IMReadColor)
	for i := 0; i < len(data.Data); i++ {
		gocv.Rectangle(&srcImage, data.Data[i], color.RGBA{R: 255}, 2)
	}
	window := gocv.NewWindow("match-result")
	window.IMShow(srcImage)
	//fmt.Println(data.Data, data.Msg)
}
