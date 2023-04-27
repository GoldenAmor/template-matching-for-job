package images

import (
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"sync"
	"time"
)

func MatchInit(source, template string, method gocv.TemplateMatchMode) gocv.Mat {
	srcImage := gocv.IMRead(source, gocv.IMReadColor)
	tmpl := gocv.IMRead(template, gocv.IMReadColor)
	rotated90 := gocv.NewMat()
	rotated180 := gocv.NewMat()
	rotated270 := gocv.NewMat()
	gocv.Rotate(tmpl, &rotated270, gocv.Rotate90CounterClockwise)
	gocv.Rotate(tmpl, &rotated180, gocv.Rotate180Clockwise)
	gocv.Rotate(tmpl, &rotated90, gocv.Rotate90Clockwise)
	resultImage := srcImage
	rectChan := make(chan image.Rectangle)
	var wg sync.WaitGroup
	wg.Add(1)
	//go func() {
	//	templateMatch(srcImage, tmpl, rectChan, method)
	//
	//	wg.Done()
	//}()
	//go func() {
	//	templateMatch(srcImage, rotated90, rectChan, method)
	//	wg.Done()
	//}()
	go func() {
		templateMatch(srcImage, rotated180, rectChan, method)
		wg.Done()

	}()
	//go func() {
	//	templateMatch(srcImage, rotated270, rectChan, method)
	//	wg.Done()
	//
	//}()
loop:
	for {
		select {
		case r := <-rectChan:
			updateImage(&resultImage, r)
		case <-time.After(3 * time.Second):
			break loop
		}
	}
	wg.Wait()
	return resultImage
}

func templateMatch(srcImage, tmpl gocv.Mat, rectChan chan image.Rectangle, method gocv.TemplateMatchMode) {
	result := gocv.NewMatWithSize(srcImage.Rows()-tmpl.Rows()+1, srcImage.Cols()-tmpl.Cols()+1, gocv.MatTypeCV32F)
	var temp float32 = 0
	for {
		gocv.MatchTemplate(srcImage, tmpl, &result, method, gocv.NewMat())
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
		fmt.Println(maxVal)
		if maxVal < 2e+08 {
			return
		}
		if maxVal == temp {
			break
		}
		r := image.Rectangle{
			Min: maxLoc,
			Max: image.Point{X: maxLoc.X + tmpl.Cols(), Y: maxLoc.Y + tmpl.Rows()},
		}
		rectChan <- r
		srcImage.SetDoubleAt(maxLoc.X, maxLoc.Y, 0)
		temp = maxVal
	}
}

func updateImage(resultImage *gocv.Mat, r image.Rectangle) {
	gocv.Rectangle(resultImage, r, color.RGBA{R: 255}, 2)
}
