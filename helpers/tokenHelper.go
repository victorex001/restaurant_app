package helpers

import (
	"github.com/dgrijalva/jwt-go" // Corrected import path
)

type SignedDetails struct {
	Email      string
	FirstName  string
	LastName   string
	Uid        string
	jwt.StandardClaims
}

func GenerateAllTokens(){

}

func UpdateAllToken(){

}


func ValidateToken(){
	
}