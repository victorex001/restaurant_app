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
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

var validate = validator.New()


func GetMenus() gin.HandlerFunc{
	return func(c *gin.Context){
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err!= nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the menu items"})
			return
		}
		var allMenus []bson.M
		if err = result.All(ctx, &allMenus); err!= nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allMenus)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		menuId := c.Param("menu_id")
		var menu models.Menu

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err!= nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the menu"})
		}
		c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()


		result, insertErr := menuCollection.InsertOne(ctx, menu)
		if insertErr != nil{
			msg := fmt.Sprintf("Menu item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
		defer cancel()
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}


func UpdateMenu(menuCollection *mongo.Collection) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        var menu models.Menu
        if err := c.BindJSON(&menu); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        menuId := c.Param("menu_id")
        filter := bson.M{"menu_id": menuId}

        updateObj := primitive.D{}
        if menu.Start_Date != nil && menu.End_Date != nil {
            updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
            updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})
        }
        if menu.Name != "" {
            updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
        }
        if menu.Category != "" {
            updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
        }
        menu.Updated_at = time.Now()
        updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

		upsert := true
        opts := options.UpdateOptions{Upsert: &upsert}

        result, err := menuCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, &opts)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, result)
    }
}
