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
	ID                     uint     `json:"id"`
	Name                   string   `json:"name"`
	PartNumber             string   `json:"part_number"`
	PartDescription        string   `json:"part_description"`
	SuperSSNumber          string   `json:"super_ss_number"`
	Weight                 string   `json:"weight"`
	HsCode                 string   `json:"hs_code"`
	RemainPartNumber       string   `json:"remain_part_number"`
	Coo                    string   `json:"coo"`
	RefNO                  string   `json:"ref_no"`
	Image                  string   `json:"image"`
	Images                 []string `json:"images"`
	MainCategory           string   `json:"main_category"`
	SubCategory            string   `json:"sub_category"`
	Dimension              string   `json:"dimension"`
	CompatibleEngineModels string   `json:"compatible_engine_models"`
	AvailableLocation      string   `json:"available_location"`
	Price                  float64  `json:"price"`
}

type HomeCompanySlides struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	CreatedDate time.Time `json:"created_date"`
}

type EnquiresModel struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Enquiry     string    `json:"enquiry"`
	Attachments string    `json:"attachment"`
	CreatedDate time.Time `json:"created_date"`
}

type Engine struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	PartNumber        string    `json:"part_number"`
	Hz                string    `json:"hz"`
	EpOrInd           string    `json:"ep_or_ind"`
	Weight            string    `json:"weight"`
	Coo               string    `json:"coo"`
	Image             string    `json:"image"`
	Description       string    `json:"description"`
	AvailableLocation string    `json:"available_location"`
	KVA               string    `json:"kva"`
	SpecificationURL  string    `json:"specification_url"`
	MainCategory      string    `json:"main_category"`
	CreatedDate       time.Time `json:"created_date"`
}

type Catalogue struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	PartNumber   string    `json:"part_number"`
	Description  string    `json:"description"`
	MainCategory string    `json:"main_category"`
	Image        string    `json:"image"`
	PdfURL       string    `json:"pdf_url"`
	CreatedDate  time.Time `json:"created_date"`
}
