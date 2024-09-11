package helpers

import (
	"context"
	"log"
	"os"
	"restaurant_app/database"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	FirstName  string
	LastName   string
	Uid        string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")


func GenerateAllTokens(email, firstName, lastName, uid string) (signedToken string, signedRefreshToken string, err error) {
    // Setup claims for the access token
    claims := &SignedDetails{
        Email:      email,
        FirstName:  firstName,
        LastName:   lastName,
        Uid:        uid,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(15 * time.Minute).Unix(), // 15 minutes for access token
        },
    }

    // Setup claims for the refresh token
    refreshClaims := &SignedDetails{
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days for refresh token
        },
    }

    // Generate the access token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signedToken, err = token.SignedString([]byte(SECRET_KEY))
    if err != nil {
        log.Printf("Failed to sign the access token: %v", err)
        return
    }

    // Generate the refresh token
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    signedRefreshToken, err = refreshToken.SignedString([]byte(SECRET_KEY))
    if err != nil {
        log.Printf("Failed to sign the refresh token: %v", err)
        return
    }
    return signedToken, signedRefreshToken, nil
}


func UpdateAllToken(signedToken string, signedRefreshToken string, userId string){
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})
	

	filter := bson.M{"user_id": userId}
	upsert := true
    opts := options.UpdateOptions{Upsert: &upsert}


	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opts,
	)
	if err != nil {
		log.Printf("Failed to update tokens: %v", err)
        return
	}
}


func ValidateToken(signedToken string) (claims *SignedDetails, msg string){
	
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})
	if err != nil{
		log.Printf("Failed to sign the refresh token: %v", err)
        return

	}


	// Check if the token is inValid
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is inValid"
		log.Printf("Failed to sign the refresh token: %v", msg)
        return
	}
	
	// Check if the token is expired
	if claims.ExpiresAt < time.Now().Local().Unix(){
		msg = "Token is expired"
		log.Printf("Failed to sign the refresh token: %v", msg)
		return
	}
	return claims, msg
}