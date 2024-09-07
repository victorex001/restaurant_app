package controller

import (
	"context"
	"log"
	"net/http"
	"restaurant_app/database"
	"restaurant_app/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderItemPack struct{
	Table_id *string
	Order_items []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing order items"})
			return
		}
		var allOrderItems []bson.M
		err = result.All(ctx, &allOrderItems)
		if err != nil{
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderItemId := c.Param("order_item_id")
		var orderItem models.OrderItem

		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while lisitng ordered item"})
			return
		}
		c.JSON(http.StatusOK, orderItem)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		orderId := c.Param("order_id")
		
		allOrderItems, err := ItemsByOrder(orderId)

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing order items by order ID"})
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}


func ItemsByOrder(id string) (OrderItems []primitive.M, err error){
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "order_id", Value: id}}}}
	lookupFoodStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "food"},
		{Key: "localField", Value: "food_id"},
		{Key: "foreignField", Value: "food_id"},
		{Key: "as", Value: "food"},
	}}}
	unwindFoodStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$food"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "order"},
		{Key: "localField", Value: "order_id"},
		{Key: "foreignField", Value: "order_id"},
		{Key: "as", Value: "order"},
	}}}
	unwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$order"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}

	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "table"},
		{Key: "localField", Value: "order.table_id"},
		{Key: "foreignField", Value: "order.table_id"},
		{Key: "as", Value: "table"}
	}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{
		{Key: "path", Value: "$table"},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}}


	projectStage := bson.D{{Key: "$project", Value: bson.M{
		"id": 0,
		"amount": "$food.price",
		"total_count": 1,
		"food_name": "$food.name",
		"food_image": "$food.food_image",
		"table_number": "$table.table_number",
		"table_id": "$table.table_id",
		"order_id": "$order.order_id",
		"quantity": 1, // Corrected from "quanitity"
	}}}
	
	groupStage := bson.D{
		{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"order_id": "$order_id",
				"table_id": "$table_id",
				"table_number": "$table_number",
			},
			"payment_due": bson.M{"$sum": "$amount"},
			"total_count": bson.M{"$sum": 1},
			"order_items": bson.M{"$push": "$$ROOT"},  // Correct usage
		}},
	}
	
	
	projectStage2 := bson.D{{Key: "$project", Value: bson.M{
		"id": 0,
		"payment_due": 1,
		"total_count": 1,
		"table_number": "$_id.table_number",
		"order_items": 1,
	}}}
	

	// Combine all stages into a pipeline

	pipeline := mongo.Pipeline{
		matchStage,
		lookupFoodStage,
		unwindFoodStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	}

	result, err := orderCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err  // Handle error appropriately
	}
	defer result.Close(ctx)  // Ensure cursor is closed

	var orderItems []primitive.M
	err = result.All(ctx, &orderItems)
	if err != nil {
		return nil, err  // Handle error appropriately
	}

	return orderItems, nil

}


func CreateOrderItem() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItemPack orderItemPack 
		var order models.Order 
		
		err := c.BindJSON(&orderItemPack)
		if err!= nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_Table_id = orderItemPack.Table_id
		order_id := orderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.Order_items{
			orderItem.Order_id = order_id

			validationErr := validate.Struct(orderItem)
			if validationErr != nil{
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at = time.Now()
			orderItem.Updated_at = time.Now() // Directly assign the current time

			orderItem.Order_item_id = orderItem.ID.Hex()

			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)

		}
		insertedOrderItems, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)

		if err != nil{
			log.Fatal(err)
		}

		c.JSON(http.StatusOk, insertedOrderItems)
	}
}

func UpdateOrderItem() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem = models.OrderItem

		orderItemId := c.Param("order_item_id")

		filter := bson.M{"order_item_id": orderItemId}

		var updateObj primitive.D

		if orderItem.Unit_price != nil{
			updateObj = append(updateObj, bson.E{Key: "unit_price", Value: *&ordrItem.Unit_price})
		}

		if orderItem.Quantity != nil{
			updateObj = append(updateObj, bson.E{Key: "quantity", Value: *orderItem.Quantity})
		}

		if orderItem.Food_id != nil{
			updateObj = append(updateObj, bson.E{Key: "food_id", Value: *orderItem.Food_id})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: orderItem.Updated_at})

		upsert := true
        opts := options.UpdateOptions{Upsert: &upsert}

		result, err := orderItemCollection.UpdateOne(
			ctx,
			filtet,
			bson.D{{Key: "$set", Value: updateObj}}, &opts
		)
		if err != nil {
			msg := "Order Item updated failed"
            c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
            return
        }
        c.JSON(http.StatusOK, result)
	}
}