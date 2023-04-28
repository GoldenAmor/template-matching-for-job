package main

import (
	"github.com/gin-gonic/gin"
	"template-matching/match/handler"
)

func RoutesController() *gin.Engine {
	Engine := gin.Default()
	Engine.GET("/template-match-serial", handler.TemplateMatchSerial)
	Engine.GET("/template-match-concurrent", handler.TemplateMatchConcurrent)
	return Engine
}
