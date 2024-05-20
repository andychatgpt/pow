package controllers

import (
	"github.com/gin-gonic/gin"
	"pow/models"
	"pow/services"
)

func Get(c *gin.Context) {
	var jsons models.ParamGetPow
	if err := c.BindJSON(&jsons); err != nil {
		c.JSON(400, gin.H{
			"code":    "fail",
			"data":    "",
			"message": err.Error(),
		})
		return
	}

	if token := services.CalcProofToken(&jsons); token != "" {
		c.JSON(200, gin.H{
			"code":    "ok",
			"data":    token,
			"message": "success",
		})
		return
	}

	c.JSON(400, gin.H{
		"code":    "fail",
		"data":    "",
		"message": "",
	})
}
