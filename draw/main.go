package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

var url string
var imagePath string

func init() {
	flag.StringVar(&url, "method", "", "指定调用同步/异步接口")
	flag.StringVar(&url, "image", "./images/src1.jpg", "指定绘制标号的图像路径")
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
	fmt.Println(string(body))
	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}
	srcImage := gocv.IMRead(imagePath, gocv.IMReadColor)
	var point image.Point
	for i := 0; i < len(data.Data); i++ {
		point = image.Point{
			X: ((data.Data[i].Max.X - data.Data[i].Min.X) / 2) + data.Data[i].Min.X,
			Y: ((data.Data[i].Max.Y - data.Data[i].Min.Y) / 2) + data.Data[i].Min.Y}
		gocv.Circle(&srcImage, point, 50, color.RGBA{255, 255, 255, 0}, 2)
		gocv.PutText(&srcImage, strconv.Itoa(i+1), point, gocv.FontHersheyPlain, 4.0, color.RGBA{255, 255, 255, 0}, 2)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(srcImage gocv.Mat) {
		runtime.LockOSThread()
		window := gocv.NewWindow("match-result")
		window.IMShow(srcImage)
		gocv.WaitKey(0)
		runtime.UnlockOSThread()
		wg.Done()
	}(srcImage)
	wg.Wait()
	fmt.Println(data.Data, data.Msg)
}

func draw(srcImage gocv.Mat) {
	window := gocv.NewWindow("match-result")
	window.IMShow(srcImage)
	gocv.WaitKey(0)
}
