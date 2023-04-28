package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
	"image"
	"sync"
	"time"
)

func TemplateMatchSerial(c *gin.Context) {
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
	var wg sync.WaitGroup
	wg.Add(4)
	result := matchLogic(srcImageGray, tmplGray, gocv.TmSqdiff)
	result90 := matchLogic(srcImageGray, rotated90, gocv.TmSqdiff)
	result180 := matchLogic(srcImageGray, rotated180, gocv.TmSqdiff)
	result270 := matchLogic(srcImageGray, rotated270, gocv.TmSqdiff)
	result = append(result, result90...)
	result = append(result, result180...)
	result = append(result, result270...)
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "successfully matched",
	})
}

func TemplateMatchConcurrent(c *gin.Context) {
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
	var tmplList []gocv.Mat
	tmplList = append(tmplList, tmplGray)
	tmplList = append(tmplList, rotated90)
	tmplList = append(tmplList, rotated180)
	tmplList = append(tmplList, rotated270)
	var wg sync.WaitGroup
	wg.Add(4)
	rectChan := make(chan image.Rectangle)
	for i := 0; i < len(tmplList); i++ {
		go func(template gocv.Mat) {
			result := matchLogic(srcImageGray, template, gocv.TmSqdiff)
			for _, v := range result {
				rectChan <- v
			}
			wg.Done()
		}(tmplList[i])
	}
	var result []image.Rectangle
loop:
	for {
		select {
		case r := <-rectChan:
			result = append(result, r)
		case <-time.After(3 * time.Second):
			break loop
		}
	}

	wg.Wait()
	close(rectChan)
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "successfully matched",
	})
}

func matchLogic(srcImage, tmpl gocv.Mat, method gocv.TemplateMatchMode) (ret []image.Rectangle) {
	result := gocv.NewMatWithSize(srcImage.Rows()-tmpl.Rows()+1, srcImage.Cols()-tmpl.Cols()+1, gocv.MatTypeCV32F)
	for {
		gocv.MatchTemplate(srcImage, tmpl, &result, method, gocv.NewMat())
		minVal, _, minLoc, _ := gocv.MinMaxLoc(result)
		fmt.Println(minVal)
		if minVal > 500000 {
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
