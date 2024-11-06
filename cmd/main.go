package main

import (
	"net/http"

	"morris-backend.com/main/database"
	"morris-backend.com/main/middleware"
	"morris-backend.com/main/services/handler"
)

func main() {
	database.Initdb()

	//part router
	http.Handle("/part", middleware.AuthMiddleware(http.HandlerFunc(handler.PartHandler)))
	http.Handle("/parts", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPartHandlerByPartNumber)))

	//Banner router
	http.Handle("/banner", middleware.AuthMiddleware(http.HandlerFunc(handler.BannerHandler)))

	//Company router
	http.Handle("/company", middleware.AuthMiddleware(http.HandlerFunc(handler.CompanyHandler)))

	//Category router
	http.Handle("/category", middleware.AuthMiddleware(http.HandlerFunc(handler.CategoryHandler)))

	http.ListenAndServe(":8080", nil)

}
