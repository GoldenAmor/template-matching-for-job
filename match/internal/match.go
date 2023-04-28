package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"sync"
	"time"
)

func MatchConcurrent(source, template string, method gocv.TemplateMatchMode) gocv.Mat {
	srcImage := gocv.IMRead(source, gocv.IMReadColor)
	tmpl := gocv.IMRead(template, gocv.IMReadColor)
	srcImageGray := gocv.NewMat()
	tmplGray := gocv.NewMat()
	gocv.CvtColor(srcImage, &srcImageGray, gocv.ColorBGRToGray)
	gocv.CvtColor(tmpl, &tmplGray, gocv.ColorBGRToGray)
	rotated90 := gocv.NewMat()
	rotated180 := gocv.NewMat()
	rotated270 := gocv.NewMat()
	src90 := gocv.NewMat()
	src180 := gocv.NewMat()
	src270 := gocv.NewMat()
	gocv.Rotate(tmplGray, &rotated270, gocv.Rotate90CounterClockwise)
	gocv.Rotate(tmplGray, &rotated180, gocv.Rotate180Clockwise)
	gocv.Rotate(tmplGray, &rotated90, gocv.Rotate90Clockwise)
	gocv.Rotate(srcImageGray, &src270, gocv.Rotate90CounterClockwise)
	gocv.Rotate(srcImageGray, &src180, gocv.Rotate180Clockwise)
	gocv.Rotate(srcImageGray, &src90, gocv.Rotate90Clockwise)
	resultImage := srcImage.Clone()
	rectChan := make(chan image.Rectangle)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		templateMatch(srcImageGray, tmplGray, rectChan, method)
		wg.Done()
	}()
	//go func() {
	//	templateMatch(srcImageGray, rotated90, rectChan, method)
	//	wg.Done()
	//}()
	//go func() {
	//	templateMatch(srcImageGray, rotated180, rectChan, method)
	//	wg.Done()
	//
	//}()
	go func() {
		templateMatch(srcImageGray, rotated270, rectChan, method)
		wg.Done()
	}()
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

func MatchSerial(c *gin.Context) {
	srcImage := gocv.IMRead("./images/src1.jpg", gocv.IMReadColor)
	tmpl := gocv.IMRead("./images/aim.png", gocv.IMReadColor)
	srcImageGray := gocv.NewMat()
	tmplGray := gocv.NewMat()
	gocv.CvtColor(srcImage, &srcImageGray, gocv.ColorBGRToGray)
	gocv.CvtColor(tmpl, &tmplGray, gocv.ColorBGRToGray)
	rotated90 := gocv.NewMat()
	rotated180 := gocv.NewMat()
	rotated270 := gocv.NewMat()
	gocv.Rotate(tmplGray, &rotated270, gocv.Rotate90CounterClockwise)
	gocv.Rotate(tmplGray, &rotated180, gocv.Rotate180Clockwise)
	gocv.Rotate(tmplGray, &rotated90, gocv.Rotate90Clockwise)
	result := templateMatchSerial(srcImageGray, tmplGray, gocv.TmCcoeff)
	result90 := templateMatchSerial(srcImageGray, rotated90, gocv.TmCcoeff)
	result180 := templateMatchSerial(srcImageGray, rotated180, gocv.TmCcoeff)
	result270 := templateMatchSerial(srcImageGray, rotated270, gocv.TmCcoeff)
	result = append(result, result90...)
	result = append(result, result180...)
	result = append(result, result270...)
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "successfully matched",
	})
}

func templateMatch(srcImage, tmpl gocv.Mat, rectChan chan image.Rectangle, method gocv.TemplateMatchMode) {
	result := gocv.NewMatWithSize(srcImage.Rows()-tmpl.Rows()+1, srcImage.Cols()-tmpl.Cols()+1, gocv.MatTypeCV32F)
	for {
		gocv.MatchTemplate(srcImage, tmpl, &result, method, gocv.NewMat())
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
		fmt.Println(maxVal)
		if maxVal < 7.5e+07 {
			return
		}
		r := image.Rectangle{
			Min: maxLoc,
			Max: image.Point{X: maxLoc.X + tmpl.Cols(), Y: maxLoc.Y + tmpl.Rows()},
		}
		rectChan <- r
		region := srcImage.Region(r)
		region.SetTo(gocv.NewScalar(255, 0, 0, 0))
	}
}

func templateMatchSerial(srcImage, tmpl gocv.Mat, method gocv.TemplateMatchMode) (ret []image.Rectangle) {
	result := gocv.NewMatWithSize(srcImage.Rows()-tmpl.Rows()+1, srcImage.Cols()-tmpl.Cols()+1, gocv.MatTypeCV32F)
	for {
		gocv.MatchTemplate(srcImage, tmpl, &result, method, gocv.NewMat())
		minVal, _, minLoc, _ := gocv.MinMaxLoc(result)
		fmt.Println(minVal)
		if minVal > -2.4e+07 {
			return
		}

		r := image.Rectangle{
			Min: minLoc,
			Max: image.Point{X: minLoc.X + tmpl.Cols(), Y: minLoc.Y + tmpl.Rows()},
		}
		//gocv.Rectangle(output, r, color.RGBA{R: 255}, 10)
		region := srcImage.Region(r)
		region.SetTo(gocv.NewScalar(255, 0, 0, 0))
		srcImage.SetDoubleAt(minLoc.X, minLoc.Y, 0)
		ret = append(ret, r)
	}
}

func updateImage(resultImage *gocv.Mat, r image.Rectangle) {
	gocv.Rectangle(resultImage, r, color.RGBA{R: 255}, 10)
}
