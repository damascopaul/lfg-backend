package endpoints

import "github.com/damascopaul/lfg-backend/schemas"

var (
	BodyInternalServerError = schemas.BodyError{
		Message: "An internal error occurred in the server"}
	BodyNotFound = schemas.BodyError{
		Message: "The requested resource could not be found"}
)
