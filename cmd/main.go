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

	http.Handle("/morrispartssearch", middleware.AuthMiddleware(http.HandlerFunc(handler.MorrisPartsSearchHandler)))

	http.Handle("/admin/parts", middleware.AuthMiddleware(http.HandlerFunc(handler.AdminPartHandler)))

	http.Handle("/admin/subcategory", middleware.AuthMiddleware(http.HandlerFunc(handler.AdminSubCategoryHandler)))

	http.Handle("/parts/home", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPartsByOnlyCategoryHandler)))

	http.Handle("/subcategory", middleware.AuthMiddleware(http.HandlerFunc(handler.GetHomeSubCategoriesHandler)))

	http.Handle("/part/subcategory", middleware.AuthMiddleware(http.HandlerFunc(handler.GetPartsByOnlySubCategoryHandler)))

	http.Handle("/enquiries", middleware.AuthMiddleware(http.HandlerFunc(handler.EnquiryHandler)))

	http.Handle("/morrisparts/part", middleware.AuthMiddleware(http.HandlerFunc(handler.PartByIDHandler)))

	http.Handle("/morrisparts/relatedparts", middleware.AuthMiddleware(http.HandlerFunc(handler.GetRelatedParts)))

	http.Handle("/engines", middleware.AuthMiddleware(http.HandlerFunc(handler.EngineHandler)))

	http.Handle("/catalogues", middleware.AuthMiddleware(http.HandlerFunc(handler.CatalogeHandler)))

	http.Handle("/engines/detail", middleware.AuthMiddleware(http.HandlerFunc(handler.EngineHandlerByID)))

	http.Handle("/catalogues/detail", middleware.AuthMiddleware(http.HandlerFunc(handler.CatalogueHandlerByID)))

	http.Handle("/customerdetails", middleware.AuthMiddleware(http.HandlerFunc(handler.CustomerDetails)))

	http.ListenAndServe(":8080", nil)

}
