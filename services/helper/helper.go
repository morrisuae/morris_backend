package helper

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"morris-backend.com/main/services/models"
)

var DB *sql.DB

// Parts GET, POST, PUT and DELETE
func PostPart(part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category string) (uint, error) {
	// Connect to the database
	var id uint

	err := DB.QueryRow("INSERT INTO parts (part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id", part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetPart() ([]models.Part, error) {
	rows, err := DB.Query("SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category FROM parts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.Part
	for rows.Next() {
		var part models.Part
		err := rows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode, &part.Image, &part.SubCategory)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	fmt.Println("Get Successful")

	return parts, nil
}

func GetPartByPartNumber(partNumber string) ([]models.Part, error) {
	var parts []models.Part

	var query string
	var args []interface{}

	if len(partNumber) >= 3 {
		// Check if the full part number is provided
		fullPartQuery := "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category FROM parts WHERE part_number = $1"
		fullPartRows, err := DB.Query(fullPartQuery, partNumber)
		if err != nil {
			log.Println("Error executing full part query:", err)
			return nil, err
		}
		defer fullPartRows.Close()

		// Check for exact match
		for fullPartRows.Next() {
			var part models.Part
			err := fullPartRows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode, &part.Image, &part.SubCategory)
			if err != nil {
				log.Println("Error scanning row:", err)
				return nil, err
			}
			parts = append(parts, part)
		}

		if len(parts) > 0 {
			// If exact match is found, return these results
			return parts, nil
		}

		// If no exact match, perform prefix search
		query = "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category FROM parts WHERE part_number LIKE $1"
		args = append(args, partNumber[:3]+"%")
	} else {
		// Handle as an exact match if partNumber is shorter than 3 characters
		query = "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category FROM parts WHERE part_number = $1"
		args = append(args, partNumber)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		log.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var part models.Part
		err := rows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode, &part.Image, &part.SubCategory)
		if err != nil {
			log.Println("Error scanning row:", err)
			return nil, err
		}
		parts = append(parts, part)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error processing rows:", err)
		return nil, err
	}

	return parts, nil
}

func PutPart(id uint, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category string) error {
	result, err := DB.Exec("UPDATE parts SET part_number=$1, remain_part_number=$2, part_description=$3, fg_wison_part_number=$4, super_ss_number=$5, weight=$6, coo=$7, hs_code=$8, image=$9, sub_category=$10 WHERE id=$11", part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category, id)

	if err != nil {
		return fmt.Errorf("failed to query part: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("part not found")
	}

	fmt.Println("Update successfull")

	return nil
}

func DeletePart(id uint) error {
	result, err := DB.Exec("DELETE FROM parts WHERE id=$1", id)

	if err != nil {
		return fmt.Errorf("failed to delete part: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("part not found")
	}

	fmt.Println("Delete successfull")

	return nil
}

// Banner GET and POST
func PostBanner(image string, created_date time.Time) error {
	// Connect to the database

	currentTime := time.Now()
	_, err := DB.Exec("INSERT INTO banners (image, created_date) VALUES ($1, $2)", image, currentTime)
	if err != nil {
		return err
	}

	fmt.Println("Post Successful")

	return nil
}

func GetBanner() ([]models.Banner, error) {
	rows, err := DB.Query("SELECT image, created_date FROM banners")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Banners []models.Banner
	for rows.Next() {
		var Banner models.Banner
		err := rows.Scan(&Banner.Image, &Banner.CreatedDate)
		if err != nil {
			return nil, err
		}
		Banners = append(Banners, Banner)
	}

	fmt.Println("Get Successful")

	return Banners, nil
}

// Company GET and POST
func PostCompany(company_name, cover_image string, created_date time.Time) (uint, error) {
	var id uint

	currentTime := time.Now()
	err := DB.QueryRow("INSERT INTO company (company_name, created_date, cover_image) VALUES ($1, $2, $3)", company_name, currentTime, cover_image).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetCompany() ([]models.Company, error) {
	rows, err := DB.Query("SELECT id, company_name, created_date, cover_image FROM company")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Companys []models.Company
	for rows.Next() {
		var Company models.Company
		err := rows.Scan(&Company.ID, &Company.CompanyName, &Company.CreatedDate, &Company.CoverImage)
		if err != nil {
			return nil, err
		}
		Companys = append(Companys, Company)
	}

	fmt.Println("Get Successful")

	return Companys, nil
}

// PartCategory GET and POST
func PostPartCategory(product_id, product_category, image string, created_date time.Time) (uint, error) {
	var id uint

	currentTime := time.Now()
	err := DB.QueryRow("INSERT INTO part (product_id, product_category, image, created_date) VALUES ($1, $2, $3, $4)", product_id, product_category, image, currentTime).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetPartCategory() ([]models.PartCategory, error) {
	rows, err := DB.Query("SELECT id, product_id, product_category, image, created_date FROM part")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Categories []models.PartCategory
	for rows.Next() {
		var Category models.PartCategory
		err := rows.Scan(&Category.ID, &Category.ProductId, &Category.ProductCategory, &Category.CreatedDate)
		if err != nil {
			return nil, err
		}
		Categories = append(Categories, Category)
	}

	fmt.Println("Get Successful")

	return Categories, nil
}

// Category POST, GET, PUT and DELETE
func PostCategory(name, category_name string, created_date time.Time) (uint, error) {
	// Connect to the database

	var id uint

	currentTime := time.Now()
	err := DB.QueryRow("INSERT INTO category (name, category_name, created_date) VALUES ($1, $2) RETURNING id", name, category_name, currentTime).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetCategory() ([]models.Category, error) {
	rows, err := DB.Query("SELECT id, name, category_name, created_date FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var category []models.Category
	for rows.Next() {
		var Categories models.Category
		err := rows.Scan(&Categories.ID, &Categories.Name, &Categories.CategoryName, Categories.CreatedDate)
		if err != nil {
			return nil, err
		}
		category = append(category, Categories)
	}

	fmt.Println("Get Successful")

	return category, nil
}

func PutCategory(id uint, name, category_name string) error {
	result, err := DB.Exec("UPDATE category SET name=$1, category_name=$2  WHERE id=$3", name, category_name, id)

	if err != nil {
		return fmt.Errorf("failed to query category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	fmt.Println("Update successfull")

	return nil
}

func DeleteCategory(id uint) error {
	result, err := DB.Exec("DELETE FROM category WHERE id=$1", id)

	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	fmt.Println("Delete successfull")

	return nil
}

// SubCategory POST, GET, PUT and DELETE
func PostSubCategory(main_category_name, sub_category_name, image string, created_date time.Time) (uint, error) {
	// Connect to the database

	var id uint

	currentTime := time.Now()
	err := DB.QueryRow("INSERT INTO subcategory (main_category_name, sub_category_name, image, created_date) VALUES ($1, $2) RETURNING id", main_category_name, sub_category_name, image, currentTime).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetSubCategory() ([]models.SubCategory, error) {
	rows, err := DB.Query("SELECT id, main_category_name, sub_category_name, image, created_date FROM subcategory")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var SubCategories []models.SubCategory
	for rows.Next() {
		var subcategory models.SubCategory
		err := rows.Scan(&subcategory.ID, &subcategory.MainCategoryName, &subcategory.SubCategoryName, &subcategory.Image, &subcategory.CreatedDate)
		if err != nil {
			return nil, err
		}
		SubCategories = append(SubCategories, subcategory)
	}

	fmt.Println("Get Successful")

	return SubCategories, nil
}

func PutSubCategory(id uint, main_category_name, sub_category_name, image string) error {
	result, err := DB.Exec("UPDATE subcategory SET main_category_name=$1, sub_category_name=$2, image=$3  WHERE id=$4", main_category_name, sub_category_name, image, id)

	if err != nil {
		return fmt.Errorf("failed to query category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	fmt.Println("Update successfull")

	return nil
}

func DeleteSubCategory(id uint) error {
	result, err := DB.Exec("DELETE FROM subcategory WHERE id=$1", id)

	if err != nil {
		return fmt.Errorf("failed to delete subcategory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subcategory not found")
	}

	fmt.Println("Delete successfull")

	return nil
}
