package main

import (
	"gocv.io/x/gocv"
	"template-matching/images"
)

func main() {
	resultImage := images.MatchInit("./images/src1.jpg", "./images/aim.png", gocv.TmCcoeff)
	window := gocv.NewWindow("result")
	window.IMShow(resultImage)
	gocv.WaitKey(0)
}
