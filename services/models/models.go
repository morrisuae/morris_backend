package models

import (
	"time"
)

type Part struct {
	ID                uint   `json:"id"`
	PartNumber        string `json:"part_number"`
	RemainPartNumber  string `json:"remain_part_number"`
	PartDescription   string `json:"part_description"`
	FgWisonPartNumber string `json:"fg_wison_part_number"`
	SuperSSNumber     string `json:"super_ss_number"`
	Weight            string `json:"weight"`
	Coo               string `json:"coo"`
	HsCode            string `json:"hs_code"`
	Image             string `json:"image"`
	SubCategory       string `json:"sub_category"`
}

type Banner struct {
	ID          uint      `json:"id"`
	Image       string    `json:"image"`
	Title       string    `json:"title"`
	CreatedDate time.Time `json:"created_date"`
}

type Company struct {
	ID          uint      `json:"id"`
	CompanyName string    `json:"company_name"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
	CoverImage  string    `json:"cover_image"`
}

type PartCategory struct {
	ID              uint      `json:"id"`
	ProductId       string    `json:"product_id"`
	ProductCategory string    `json:"product_category"`
	Image           string    `json:"image"`
	CreatedDate     time.Time `json:"created_date"`
	UpdatedDate     time.Time `json:"updated_date"`
}

type Category struct {
	ID           uint      `json:"id"`
	Image        string    `json:"image"`
	CategoryName string    `json:"category_name"`
	CreatedDate  time.Time `json:"created_date"`
}

type SubCategory struct {
	ID               uint      `json:"id"`
	MainCategoryName string    `json:"category_name"`
	SubCategoryName  string    `json:"Sub_category_name"`
	Image            string    `json:"image"`
	CategoryType     string    `json:"category_type"`
	CreatedDate      time.Time `json:"created_date"`
}

type MorrisParts struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	PartNumber       string `json:"part_number"`
	PartDescription  string `json:"part_description"`
	SuperSSNumber    string `json:"super_ss_number"`
	Weight           string `json:"weight"`
	HsCode           string `json:"hs_code"`
	RemainPartNumber string `json:"remain_part_number"`
	Coo              string `json:"coo"`
	RefNO            string `json:"ref_no"`
	Image            string `json:"image"`
	MainCategory     string `json:"main_category"`
	SubCategory      string `json:"sub_category"`
}

type HomeCompanySlides struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	CreatedDate time.Time `json:"created_date"`
}
