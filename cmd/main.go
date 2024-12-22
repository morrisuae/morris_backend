package main

import (
	"net/http"

	"morris-backend.com/main/database"
	"morris-backend.com/main/middleware"
	"morris-backend.com/main/services/handler"
)

func main() {
	database.Initdb()

	//Part router
	http.Handle("/part", middleware.AuthMiddleware(http.HandlerFunc(handler.PartHandler)))
	http.Handle("/parts", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPartHandlerByPartNumber)))

	//Banner router
	http.Handle("/banner", middleware.AuthMiddleware(http.HandlerFunc(handler.BannersHandler)))

	//Company router
	http.Handle("/company", middleware.AuthMiddleware(http.HandlerFunc(handler.CompanyHandler)))

	//PartCategory router
	http.Handle("/category", middleware.AuthMiddleware(http.HandlerFunc(handler.PartCategoryHandler)))

	//Category router
	http.Handle("/categories", middleware.AuthMiddleware(http.HandlerFunc(handler.CategoryHandler)))

	//Subcatefory router
	http.Handle("/subcategories", middleware.AuthMiddleware(http.HandlerFunc(handler.SubCategoryHandler)))

	http.Handle("/homesliders", middleware.AuthMiddleware(http.HandlerFunc(handler.HomeSliderBannerHandler)))
	//Subcatefory router
	http.Handle("/morrisparts", middleware.AuthMiddleware(http.HandlerFunc(handler.MorrisPartsHandler)))

	http.ListenAndServe(":8080", nil)
}
