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
	Image       string    `json:"image"`
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
	Name         string    `json:"name"`
	CategoryName string    `json:"category_name"`
	CreatedDate  time.Time `json:"created_date"`
}

type SubCategory struct {
	ID               uint      `json:"id"`
	MainCategoryName string    `json:"main_category_name"`
	SubCategoryName  string    `json:"Sub_category_name"`
	Image            string    `json:"image"`
	CreatedDate      time.Time `json:"created_date"`
}
