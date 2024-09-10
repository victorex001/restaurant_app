package controller

import (
	"context"
	"log"
	"net/http"
	"restaurant_app/database"
	"restaurant_app/models"
	"restaurant_app/helpers"
	"strconv"
	"time"
	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate validator.New()


func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Handling pagination parameters
		recordPerPage, err := strconv.Atoi(c.DefaultQuery("recordPerPage", "10"))
		if err != nil {
			recordPerPage = 10 // default value
		}

		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil {
			page = 1 // default value
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}} // Adjust this to filter your documents if needed


        // Projection stage to shape the output, including a pagination mechanism on a field "data"
        projectStage := bson.D{
            {Key: "$project", Value: bson.D{
                {Key: "_id", Value: 0},
                {Key: "total_count", Value: 1}, // Include total count in the output
                {Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}}, // Paginate "data" array
            }}}
        // Build the pipeline with all stages
        pipeline := mongo.Pipeline{matchStage, projectStage}

		result, err := userCollection.Aggregate(ctx, pipeline)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute query:"})
            return
        }

        var allUsers []bson.M
        if err = result.All(ctx, &allUsers); err != nil {
            log.Fatal(err)
            return
        }

        c.JSON(http.StatusOK, allUsers[0])

	}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userId := c.Param("user_id")

		var user models.User
		
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}


func SignUp() gin.HandlerFunc{
	return func(c *gin.Context){

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		// Convert the JSON data coming from postman to something golang can understand 
		err := c. BindJSON(&user)
		if err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Validate the data based on your Struct
		validationErr := validate.Struct(user)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error", validationErr.Error()})
			return
		}
		// You'll check if the email has already been used by another user

		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing email"})
			return
		}
		if emailCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email already in use"})
			return
		}

		// Hash password
		password := HashPassword(*user.Password)
		user.Password = &password


		// You'll also check if the phone number has already been used
		phoneCount, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.phone})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing phone numbers"})
			return
		}
		if phoneCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone number already in use"})
			return
		}
		
		// Create some extra details for the user object - created_at, updated_at, ID
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID
		user.User_id = user.ID.Hex()

		
		// Generate token and refresh token (generate all tokens functions from helper)
		token, refreshToken, _ := helper.GenerateAllToken(*user.Email, *user.First_name, *user.Last_name, *user.User_id)
		user.Token = &token
		user.Refresh_Token = &refresh

		// If all ok, then you have insert this user into the user collection
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// Return status OK and send the result back
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        var user, foundUser models.User

        // Convert the login data from JSON to a Go struct
        if err := c.BindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
            return
        }

        // Find the user by email
        err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or login incorrect"})
            return
        }

        // Verify the password
        passwordIsValid, msg := VerifyPassword(user.Password, foundUser.Password)
		if !passwordIsValid {
			
            c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
            return
        }

        // Generate tokens
        token, refreshToken, err := helper.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.UserID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
            return
        }

        // Update tokens
        helper.UpdateAllTokens(token, refreshToken, foundUser.UserID)

        // Return successful login data
        c.JSON(http.StatusOK, gin.H{
            "message": "Login successful",
            "token": token,
            "refreshToken": refreshToken,
        })
    }
}


func HashPassword(password string) string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil{
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, providedPassword string) (bool, string) {
    err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
    if err != nil {
        return false, "Login or password is incorrect"
    }
    return true, ""
}