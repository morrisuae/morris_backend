package helper

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"morris-backend.com/main/services/models"
)

var DB *sql.DB

// Parts GET, POST, PUT and DELETE
func PostPart(id uint, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category string) (uint, error) {
	_, err := DB.Exec("INSERT INTO parts (id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, image, sub_category)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful with Manual ID:", id)
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

func GetBanner() ([]models.Banner, error) {
	rows, err := DB.Query("SELECT image, title, created_date FROM banners")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Banners []models.Banner
	for rows.Next() {
		var Banner models.Banner
		err := rows.Scan(&Banner.Image, &Banner.Title, &Banner.CreatedDate)
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

func PostCategory(image, category_name string, created_date time.Time) (uint, error) {

	var id uint

	// Set the current time for created_date if not provided
	if created_date.IsZero() {
		created_date = time.Now()
	}

	// Insert the category into the database
	err := DB.QueryRow(
		"INSERT INTO category (image, category_name, created_date) VALUES ($1, $2, $3) RETURNING id",
		image, category_name, created_date,
	).Scan(&id)

	// Handle potential errors during insertion
	if err != nil {
		return 0, fmt.Errorf("failed to insert category: %w", err)
	}

	fmt.Println("PostCategory Successful")
	return id, nil
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

func GetCategory() ([]models.Category, error) {
	// Query the database to get all categories
	rows, err := DB.Query("SELECT id, image, category_name, created_date FROM category")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer rows.Close()

	// Slice to hold all retrieved categories
	var categories []models.Category

	// Iterate through the result set
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Image, &category.CategoryName, &category.CreatedDate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	fmt.Println("GetCategory Successful")
	return categories, nil
}

func GetMorrisParts() ([]models.MorrisParts, error) {
	rows, err := DB.Query(`
		SELECT id, name, part_number, part_description, super_ss_number, weight, 
		       hs_code, remain_part_number, coo, ref_no, image, main_category, sub_category 
		FROM morrisparts
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var morrisPartsList []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		err := rows.Scan(
			&part.ID,
			&part.Name,
			&part.PartNumber,
			&part.PartDescription,
			&part.SuperSSNumber,
			&part.Weight,
			&part.HsCode,
			&part.RemainPartNumber,
			&part.Coo,
			&part.RefNO,
			&part.Image,
			&part.MainCategory,
			&part.SubCategory,
		)
		if err != nil {
			return nil, err
		}
		morrisPartsList = append(morrisPartsList, part)
	}

	fmt.Println("Get Morris Parts Successful")

	return morrisPartsList, nil

}

// MorrisParts POST
func PostMorrisParts(name, part_number, part_description, super_ss_number, weight, hs_code, remain_part_number, coo, ref_no, image, main_category, sub_category string) (uint, error) {

	var id uint

	currentTime := time.Now()
	err := DB.QueryRow(`
		INSERT INTO morrisparts 
		(name, part_number, part_description, super_ss_number, weight, hs_code, remain_part_number, coo, ref_no, image, main_category, sub_category, created_date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		RETURNING id`,
		name, part_number, part_description, super_ss_number, weight, hs_code, remain_part_number, coo, ref_no, image, main_category, sub_category, currentTime,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Morris Parts Successful")

	return id, nil
}

func DeleteMorrisPart(id uint) error {
	result, err := DB.Exec("DELETE FROM morrisparts WHERE id=$1", id)

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

func CreateBanner(image, title string) (uint, error) {
	var id uint
	currentTime := time.Now()
	err := DB.QueryRow(
		"INSERT INTO banners (image, title, created_date) VALUES ($1, $2, $3) RETURNING id",
		image, title, currentTime,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func GetBanners() ([]models.Banner, error) {
	rows, err := DB.Query("SELECT id, image, title, created_date FROM banners")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []models.Banner
	for rows.Next() {
		var banner models.Banner
		err := rows.Scan(&banner.ID, &banner.Image, &banner.Title, &banner.CreatedDate)
		if err != nil {
			return nil, err
		}
		banners = append(banners, banner)
	}
	return banners, nil
}
func DeleteBanner(id uint) error {
	result, err := DB.Exec("DELETE FROM banners WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Banner not found")
	}
	return nil
}

func InsertHomeCompanySlide(name, image string) (uint, error) {
	query := `INSERT INTO home_company_slides (name, image, created_date) VALUES ($1, $2, NOW()) RETURNING id`
	var id uint
	err := DB.QueryRow(query, name, image).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetHomeCompanySlides() ([]models.HomeCompanySlides, error) {
	query := `SELECT id, name, image, created_date FROM home_company_slides ORDER BY created_date DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slides []models.HomeCompanySlides
	for rows.Next() {
		var slide models.HomeCompanySlides
		err := rows.Scan(&slide.ID, &slide.Name, &slide.Image, &slide.CreatedDate)
		if err != nil {
			return nil, err
		}
		slides = append(slides, slide)
	}
	return slides, nil
}

func DeleteHomeCompanySlide(id uint) error {
	query := `DELETE FROM home_company_slides WHERE id = ?`
	result, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("slide not found")
	}
	return nil
}

func InsertSubCategory(mainCategoryName, subCategoryName, image, categoryType string) (uint, error) {
	var id uint
	currentTime := time.Now()
	err := DB.QueryRow(
		"INSERT INTO subcategories (category_name, sub_category_name, image, category_type, created_date) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		mainCategoryName, subCategoryName, image, categoryType, currentTime,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func GetSubCategories() ([]models.SubCategory, error) {
	query := `SELECT id, category_name, sub_category_name, image, category_type, created_date FROM subcategories ORDER BY created_date DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subCategories []models.SubCategory
	for rows.Next() {
		var subCategory models.SubCategory
		err := rows.Scan(&subCategory.ID, &subCategory.MainCategoryName, &subCategory.SubCategoryName, &subCategory.Image, &subCategory.CategoryType, &subCategory.CreatedDate)
		if err != nil {
			return nil, err
		}
		subCategories = append(subCategories, subCategory)
	}

	// Check if there was an error during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subCategories, nil
}

func GetSubCategoriesByCategoryNameAndType(categoryName, categoryType string) ([]models.SubCategory, error) {
	query := `
		SELECT id, category_name, Sub_category_name, image, category_type, created_date 
		FROM subcategories 
		WHERE category_name = $1 AND category_type = $2 
		ORDER BY created_date DESC`

	rows, err := DB.Query(query, categoryName, categoryType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subCategories []models.SubCategory
	for rows.Next() {
		var subCategory models.SubCategory
		err := rows.Scan(&subCategory.ID, &subCategory.MainCategoryName, &subCategory.SubCategoryName, &subCategory.Image, &subCategory.CategoryType, &subCategory.CreatedDate)
		if err != nil {
			return nil, err
		}
		subCategories = append(subCategories, subCategory)
	}

	// Check for any iteration errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subCategories, nil
}
