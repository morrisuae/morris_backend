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
}

type Banner struct {
	Image       string    `json:image`
	CreatedDate time.Time `json:"created_date"`
}

type Company struct {
	ID          uint      `json:"id"`
	CompanyName string    `json:"company_name"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
	CoverImage  string    `json:"cover_image"`
}
