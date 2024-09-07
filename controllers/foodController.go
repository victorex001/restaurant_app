package controller

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"restaurant_app/database"
	"restaurant_app/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

var validate validator.New()


func GetFoods() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Context with timeout for database operations
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        // Parse pagination parameters with defaults
        recordPerPage, err := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
        if err != nil || recordPerPage < 1 {
            recordPerPage = 10 // Default and minimum records per page
        }

        page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
        if err != nil || page < 1 {
            page = 1 // Default and minimum page number
        }

        startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

        // MongoDB aggregation pipeline
        matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}} // Adjust this to filter your documents if needed

        // We will now use the groupStage to count documents
        groupStage := bson.D{
            {Key: "$group", Value: bson.D{
                {Key: "_id", Value: nil}, // Grouping without a specific field to count all
                {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
            }},
        }

        // Projection stage to shape the output, including a pagination mechanism on a field "data"
        projectStage := bson.D{
            {Key: "$project", Value: bson.D{
                {Key: "_id", Value: 0},
                {Key: "total_count", Value: 1}, // Include total count in the output
                {Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}, // Paginate "data" array
            }}}
        // Build the pipeline with all stages
        pipeline := mongo.Pipeline{matchStage, groupStage, projectStage}

        // Execute the aggregation query
        result, err := foodCollection.Aggregate(ctx, pipeline)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query:"})
            return
        }

        var allFoods []bson.M
        if err := result.All(ctx, &allFoods); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode results"})
            return
        }

        c.JSON(http.StatusOK, allFoods[0])
    }
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()

		if err!= nil{
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while fetching the food item"})
		}
		c.JSON(http.StatusOK, food)

	}
}

func CreateFood() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		validationErr := validate.Struct(food)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		// To confirm the menu exist before creating the food
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err!= nil{
			msg := fmt.Sprintf("menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr!= nil{
			msg := fmt.Sprintf("Food item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64{
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) /  output
}

func UpdateFood() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var menu models.Menu
		var food models.Food
		foodID := c.Param("food_id")

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updateObj := primitive.D{}

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "food_price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()
			if err!= nil{
				msg := fmt.Sprintf("message: Menu was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: food.Price})
		}
		
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

		
		filter := bson.M{"food_id": foodID}

		upsert := true
        opts := options.UpdateOptions{Upsert: &upsert}

		result, err := foodCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opts)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
		c.JSON(http.StatusOK, result)
	}
}