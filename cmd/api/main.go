package main

import (
	"os"

	"github.com/RockX-SG/frost-dkg-demo/internal/api"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	"github.com/gin-gonic/gin"
)

func main() {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8000"
	}

	h := api.New()

	r := gin.Default()
	r.GET("/ping", ping.HandlePing)
	r.GET("/data/:request_id", h.HandleGetData)
	r.GET("/deposit_data/:request_id", h.HandleGetDepositData)
	r.GET("/data/:request_id/:operator_id", h.HandleGetDataByOperatorID)
	r.POST("/keygen", h.HandleKeygen)
	r.POST("/resharing", h.HandleResharing)

	panic(r.Run(addr))
}
