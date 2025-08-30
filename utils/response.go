package utils

// ErrorResponse digunakan untuk dokumentasi Swagger response error

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Status   int         `json:"status"`
	Endpoint string      `json:"endpoint"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"` // optional data
}

type SuccessResponse struct {
	Status   int         `json:"status"`
	Endpoint string      `json:"endpoint"`
	Message  interface{} `json:"message"`
	Data     interface{} `json:"data"`
}

func SendErrorResponse(c *gin.Context, code int, message string, data ...interface{}) {
	var responseData interface{} = nil
	if len(data) > 0 {
		responseData = data[0] // ambil data pertama jika ada
	}
	c.JSON(code, ErrorResponse{
		Status:   code,
		Endpoint: c.FullPath(),
		Message:  message,
		Data:     responseData,
	})
}

func SendSuccessResponse(c *gin.Context, code int, message string, data ...interface{}) {
	var responseData interface{} = nil
	if len(data) > 0 {
		responseData = data[0] // ambil data pertama jika ada
	}
	c.JSON(code, SuccessResponse{
		Status:   code,
		Endpoint: c.FullPath(),
		Message:  message,
		Data:     responseData,
	})
}

// SuccessResponse digunakan jika kamu ingin dokumentasi swagger response success dengan format serupa
