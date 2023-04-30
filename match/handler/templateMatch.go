package handler

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"gocv.io/x/gocv"
	"image"
	"net/http"
	"sort"
	"sync"
	"time"
)

var srcPath string
var tmplPath string
var configPath string

func init() {
	flag.StringVar(&configPath, "config", "../conf/conf.ini", "指定配置文件的路径")
	flag.Parse()
	cfg, err := ini.Load(configPath)
	if err != nil {
		fmt.Printf("无法加载配置文件: %v\n", err)
		return
	}
	iniSection := cfg.Section("image")
	srcPath = iniSection.Key("src").String()
	tmplPath = iniSection.Key("tmpl").String()
}

func TemplateMatchSerial(c *gin.Context) {
	srcImage := gocv.IMRead(srcPath, gocv.IMReadColor)
	tmpl := gocv.IMRead(tmplPath, gocv.IMReadColor)
	if tmpl.Rows() > srcImage.Rows() || tmpl.Cols() > srcImage.Cols() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "模板图片大于原图片！请重新配置参数并运行服务!",
		})
		return
	}
	srcImageGray := gocv.NewMat()
	tmplGray := gocv.NewMat()
	gocv.CvtColor(srcImage, &srcImageGray, gocv.ColorBGRToGray)
	gocv.CvtColor(tmpl, &tmplGray, gocv.ColorBGRToGray)
	tmplList := rotateTemplate(tmplGray)
	var result []image.Rectangle
	for i := 0; i < len(tmplList); i++ {
		result = append(result, matchLogic(srcImageGray, tmplList[i], gocv.TmSqdiff)...)
	}
	sortTarget(result, tmpl.Rows())
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "successfully matched",
	})
}

func TemplateMatchConcurrent(c *gin.Context) {
	srcImage := gocv.IMRead(srcPath, gocv.IMReadColor)
	tmpl := gocv.IMRead(tmplPath, gocv.IMReadColor)
	if tmpl.Rows() > srcImage.Rows() || tmpl.Cols() > srcImage.Cols() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "模板图片大于原图片！请重新配置参数并运行服务!",
		})
		return
	}
	srcImageGray := gocv.NewMat()
	tmplGray := gocv.NewMat()
	gocv.CvtColor(srcImage, &srcImageGray, gocv.ColorBGRToGray)
	gocv.CvtColor(tmpl, &tmplGray, gocv.ColorBGRToGray)
	tmplList := rotateTemplate(tmplGray)
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
		case <-time.After(3 * time.Second): //超时机制：如果超过设定时间仍未有匹配目标，退出循环
			break loop
		}
	}

	wg.Wait()
	close(rectChan)
	result = sortTarget(result, tmpl.Rows())
	c.JSON(200, gin.H{
		"data": result,
		"msg":  "successfully matched",
	})
}

//matchLogic 对图像进行多次模板匹配，每次将得分最高的目标加入返回值中，并将目标区域涂白，防止混淆下次匹配结果
func matchLogic(srcImage, tmpl gocv.Mat, method gocv.TemplateMatchMode) (ret []image.Rectangle) {
	result := gocv.NewMatWithSize(srcImage.Rows()-tmpl.Rows()+1, srcImage.Cols()-tmpl.Cols()+1, gocv.MatTypeCV32F)
	for {
		gocv.MatchTemplate(srcImage, tmpl, &result, method, gocv.NewMat())
		minVal, _, minLoc, _ := gocv.MinMaxLoc(result)
		if minVal > 500000 {
			return
		}
		r := image.Rectangle{
			Min: minLoc,
			Max: image.Point{X: minLoc.X + tmpl.Cols(), Y: minLoc.Y + tmpl.Rows()},
		}
		region := srcImage.Region(r)
		region.SetTo(gocv.NewScalar(255, 0, 0, 0))
		srcImage.SetDoubleAt(minLoc.X, minLoc.Y, 0)
		ret = append(ret, r)
	}
}

//sortTarget 对模板匹配的多个目标按照从左上到右下的顺序进行排序。
func sortTarget(result []image.Rectangle, gap int) []image.Rectangle {
	sort.Slice(result, func(i, j int) bool {
		return result[i].Min.Y < result[j].Min.Y || (result[i].Min.Y == result[j].Min.Y && result[i].Min.X < result[j].Min.X)
	})
	gap = result[0].Min.Y + gap
	tempIndex := 0
	for i := 0; i < len(result); i++ {
		if result[i].Min.Y > gap {
			gap = result[i].Min.Y + gap
			tempSlice := make([]image.Rectangle, i-tempIndex)
			copy(tempSlice, result[tempIndex:i])
			result = append(result[:tempIndex], result[i:]...)
			sort.Slice(tempSlice, func(k, l int) bool {
				return tempSlice[k].Min.X < tempSlice[l].Min.X
			})
			tempSlice = append(tempSlice, result[tempIndex:]...)
			result = append(result[:tempIndex], tempSlice...)
			tempIndex = i
		}
	}
	tempSlice := make([]image.Rectangle, len(result)-tempIndex)
	copy(tempSlice, result[tempIndex:])
	sort.Slice(tempSlice, func(k, l int) bool {
		return tempSlice[k].Min.X < tempSlice[l].Min.X
	})
	result = append(result[:tempIndex], tempSlice...)
	return result
}

//rotateTemplate 返回包含4个90度旋转的模板图片数组
func rotateTemplate(tmplGray gocv.Mat) []gocv.Mat {
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
	return tmplList
}
