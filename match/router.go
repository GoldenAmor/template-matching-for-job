package main

import (
	"github.com/gin-gonic/gin"
	"template-matching/match/internal"
)

func RoutesController() *gin.Engine {
	Engine := gin.Default()
	Engine.GET("/template-match", internal.MatchSerial)
	return Engine
}
