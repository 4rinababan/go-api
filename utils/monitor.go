package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupMonitoring(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
