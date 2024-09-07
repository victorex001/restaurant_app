package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"restaurant_app/database"
	"restaurant_app/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := tableCollection.Find(context.TODO(), bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing table items"})
		}
		var allTables []bson.M 
		if err := result.All(ctx, &allTables); err != nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		tableId := c.Param("table_id")
		var table models.Table

		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the table"})
			return
		}
		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var table models.Table
		err := c.BindJSON(&table)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": validationErr.Error()})
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctx, table)
		if insertErr != nil{
			msg := fmt.Sprintf("Table Item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return 
		}
		c.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		
		var table models.Table 

		tableId := c.Param("table_id")
		filter := bson.M{"table_id": tableId}
		
		err := c.BindJSON(&table)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateObj := primitive.D{}

		if table.Number_of_guest != nil {
            updateObj = append(updateObj, bson.E{Key: "number_of_guest", Value: table.Number_of_guest})
        }

        // Check if Table_number is provided and update accordingly
        if table.Table_number != nil {
            updateObj = append(updateObj, bson.E{Key: "table_number", Value: table.Table_number})
        }


		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		upsert := true
        opts := options.UpdateOptions{Upsert: &upsert}

		result, err := tableCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opts)
		if err != nil{
			msg := fmt.Sprintf("table item updated failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}