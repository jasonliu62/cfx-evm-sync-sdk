package main

import (
	"cfx-evm-sync-sdk/controllers"
	"cfx-evm-sync-sdk/store/cfxMysql"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	//config.InitConfig()
	//nodeUrl := viper.GetStringSlice("nodes")[0]
	db := cfxMysql.Start()
	router := gin.Default()
	controllers.InitRoutes(router, db)
	// simpleBiz.ContinueBlockByNumber(nodeUrl, uint64(97971351), db)
	err := router.Run(":8080")
	if err != nil {
		log.Println("Failed to start server")
	}
}
