package controllers

import (
	"cfx-evm-sync-sdk/biz/simpleBiz"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func InitRoutes(router *gin.Engine, db *gorm.DB) {
	blockController := BlockController{DB: db}
	router.POST("/continue-block", blockController.ContinueBlockHandler)
}

type BlockController struct {
	DB *gorm.DB
}

func (bc *BlockController) ContinueBlockHandler(c *gin.Context) {
	var request struct {
		Node       string `json:"node"`
		StartBlock uint64 `json:"startBlock"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	go simpleBiz.ContinueBlockByNumber(request.Node, request.StartBlock, bc.DB)
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}
