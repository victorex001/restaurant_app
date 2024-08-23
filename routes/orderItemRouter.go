package routes

import (
	controller "restuarant_app/controllers"

	"github.com/gin-gonic/gin"
)


func OrderItemRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.GET("/orderItems", controller.GetOrderItems())
	incomingRoutes.GET("/orderItems/:orderItem_id", controller.GetOrderItem())
	incomingRoutes.GET("/orderItems-order/:order_id", controller.GetOrderItemsByOrder())
	incomingRoutes.POST("orderItems", controller.CreateOderItem())
	incomingRoutes.PATCH("/orderItems/:orderItem_id", controller.UpdateOderItem())
}