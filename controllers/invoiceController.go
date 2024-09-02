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
	"go.mongodb.org/mongo-driver/mongo"
)

type InvoiceViewFormat struct{
	Invoice_id			string
	Payment_method		string
	Order_id			string
	Payment_status		*string
	Payment_due			interface{}
	Table_number		interface{}
	Payment_due_date	time.Time
	Order_details		interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")


func GetInvoices() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := invoiceCollection.Find(context.TODO(), bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing invoice items"})
		}

		var allInvoices []bson.M
		if err = result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		
		invoiceId := c.Param("invoice_id")

		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)

		if er != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while fetching the invoice item"})
		}

		}
	}
}

func CreateInvoice() gin.HandlerFunc{
	return func(c *gin.Context){

	}
}

func UpdateInvoice() gin.HandlerFunc{
	return func(c *gin.Context){
		
	}
}