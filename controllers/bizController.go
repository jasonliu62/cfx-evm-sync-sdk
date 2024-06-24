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
	router.GET("/erc20_transfers/:address", blockController.GetErc20Transfers)
	router.POST("/check-erc20", blockController.CheckErc20)
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

func (bc *BlockController) GetErc20Transfers(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}
	transfers, err := simpleBiz.GetErc20Transfers(bc.DB, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if transfers == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	c.JSON(http.StatusOK, transfers)
}

func (bc *BlockController) CheckErc20(c *gin.Context) {
	var request struct {
		Node    string `json:"node"`
		Address string `json:"address"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	go simpleBiz.TestErc20(request.Address, request.Node)
	c.JSON(http.StatusOK, gin.H{"status": "started"})
}
