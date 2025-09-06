package helper

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
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

func PostMorrisParts(
	name, partNumber, partDescription, superSSNumber, weight, hsCode,
	remainPartNumber, coo, refNo, image, mainCategory, subCategory,
	dimension, compatibleEngineModels, availableLocation string,
	price float64, images []string,
) (uint, error) {

	var id uint
	currentTime := time.Now()

	// Ensure non-nil slice for pq.Array
	if images == nil {
		images = []string{}
	}

	query := `
        INSERT INTO morrisparts (
            name, part_number, part_description, super_ss_number,
            weight, hs_code, remain_part_number, coo, ref_no,
            image, main_category, sub_category, dimension,
            compatible_engine_models, available_location, price,
            images, created_date
        )
        VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9,
            $10, $11, $12, $13, $14, $15, $16,
            $17, $18
        )
        RETURNING id
    `

	err := DB.QueryRow(
		query,
		name, partNumber, partDescription, superSSNumber,
		weight, hsCode, remainPartNumber, coo, refNo,
		image, mainCategory, subCategory, dimension,
		compatibleEngineModels, availableLocation, price,
		pq.Array(images), currentTime,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	fmt.Println("Post Morris Parts Successful, ID:", id)
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
	result, err := DB.Exec("DELETE FROM home_company_slides WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("slide not found")
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

func GetPartsByCategory(mainCategory, subCategory string) ([]models.MorrisParts, error) {
	if mainCategory == "" || subCategory == "" {
		return nil, errors.New("main_category and sub_category are required")
	}

	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, images, main_category, sub_category,
		       dimension, compatible_engine_models, available_location, price
		FROM morrisparts
		WHERE main_category = $1 AND sub_category = $2
		ORDER BY id ASC
	`

	rows, err := DB.Query(query, mainCategory, subCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		var imagesJSON []byte // for JSONB
		var dimension, compatibleEngineModels, availableLocation sql.NullString
		var price sql.NullFloat64

		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription, &part.SuperSSNumber,
			&part.Weight, &part.HsCode, &part.RemainPartNumber, &part.Coo, &part.RefNO,
			&part.Image, &imagesJSON, &part.MainCategory, &part.SubCategory,
			&dimension, &compatibleEngineModels, &availableLocation, &price,
		)
		if err != nil {
			return nil, err
		}

		// Decode JSONB into []string
		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &part.Images); err != nil {
				return nil, fmt.Errorf("failed to unmarshal images JSON: %w", err)
			}
		}

		// Handle nullable fields
		if dimension.Valid {
			part.Dimension = dimension.String
		}
		if compatibleEngineModels.Valid {
			part.CompatibleEngineModels = compatibleEngineModels.String
		}
		if availableLocation.Valid {
			part.AvailableLocation = availableLocation.String
		}
		if price.Valid {
			part.Price = price.Float64
		} else {
			part.Price = 0
		}

		parts = append(parts, part)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

func SearchParts(partNumber string) ([]models.MorrisParts, error) {
	if partNumber == "" {
		return nil, errors.New("part_number is required")
	}

	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, main_category, sub_category
		FROM morrisparts
		WHERE part_number ILIKE $1
		ORDER BY id ASC`

	rows, err := DB.Query(query, "%"+partNumber+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription, &part.SuperSSNumber,
			&part.Weight, &part.HsCode, &part.RemainPartNumber, &part.Coo, &part.RefNO,
			&part.Image, &part.MainCategory, &part.SubCategory,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

//ADMIN

// GetMorrisParts retrieves all MorrisParts from the database.
func GetAdminParts(db *sql.DB) ([]models.MorrisParts, error) {
	query := `
		SELECT 
			id, name, part_number, part_description, super_ss_number, 
			weight, hs_code, remain_part_number, coo, ref_no, 
			image, main_category, sub_category 
		FROM morris_parts
		ORDER BY name ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription,
			&part.SuperSSNumber, &part.Weight, &part.HsCode, &part.RemainPartNumber,
			&part.Coo, &part.RefNO, &part.Image, &part.MainCategory, &part.SubCategory,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return nil, errors.New("no parts found")
	}

	return parts, nil
}

// GetSubCategories retrieves all subcategories from the database.
func GetAdminSubCategories() ([]models.SubCategory, error) {
	query := `
		SELECT 
			id, main_category_name, sub_category_name, image, category_type, created_date 
		FROM sub_categories
		ORDER BY created_date DESC
	`

	rows, err := DB.Query(query) // Assume `DB` is your database connection instance.
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subCategories []models.SubCategory
	for rows.Next() {
		var subCategory models.SubCategory
		err := rows.Scan(
			&subCategory.ID,
			&subCategory.MainCategoryName,
			&subCategory.SubCategoryName,
			&subCategory.Image,
			&subCategory.CategoryType,
			&subCategory.CreatedDate,
		)
		if err != nil {
			return nil, err
		}
		subCategories = append(subCategories, subCategory)
	}

	return subCategories, nil
}

// DeleteSubCategory deletes a subcategory by its ID.
func DeleteSubCategory(id uint) error {
	// Execute the DELETE query
	result, err := DB.Exec("DELETE FROM subcategories WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("failed to delete subcategory: %w", err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	// If no rows were affected, the subcategory was not found
	if rowsAffected == 0 {
		return fmt.Errorf("subcategory not found")
	}

	fmt.Println("Delete successful")
	return nil
}

// UpdatePart updates a part in the database based on its ID.
func UpdatePart(part models.MorrisParts) error {
	query := `
		UPDATE morrisparts 
		SET 
			name = $1, 
			part_number = $2, 
			part_description = $3, 
			super_ss_number = $4, 
			weight = $5, 
			hs_code = $6, 
			remain_part_number = $7, 
			coo = $8, 
			ref_no = $9, 
			image = $10, 
			main_category = $11, 
			sub_category = $12
		WHERE id = $13
	`

	result, err := DB.Exec(query,
		part.Name,
		part.PartNumber,
		part.PartDescription,
		part.SuperSSNumber,
		part.Weight,
		part.HsCode,
		part.RemainPartNumber,
		part.Coo,
		part.RefNO,
		part.Image,
		part.MainCategory,
		part.SubCategory,
		part.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update part: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to determine affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("part not found")
	}

	fmt.Println("Update successful")
	return nil
}

func GetPartsOnlyByCategory(mainCategory string) ([]models.MorrisParts, error) {
	if mainCategory == "" {
		return nil, errors.New("main_category is required")
	}

	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, main_category, sub_category
		FROM morrisparts
		WHERE main_category = $1
		ORDER BY id ASC`

	rows, err := DB.Query(query, mainCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription, &part.SuperSSNumber,
			&part.Weight, &part.HsCode, &part.RemainPartNumber, &part.Coo, &part.RefNO,
			&part.Image, &part.MainCategory, &part.SubCategory,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

func GetSubCategoriesWithoutQuary() ([]models.SubCategory, error) {
	// Query to fetch all subcategories, without filtering by category_name and category_type
	query := `
		SELECT id, category_name, Sub_category_name, image, category_type, created_date 
		FROM subcategories 
		ORDER BY created_date DESC`

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

	// Check for any iteration errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subCategories, nil
}

func GetPartsOnlyBySubCategory(subCategory string) ([]models.MorrisParts, error) {
	// Validate that subCategory is not empty
	if subCategory == "" {
		return nil, errors.New("sub_category is required")
	}

	// Query to fetch parts based on sub_category
	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, main_category, sub_category
		FROM morrisparts
		WHERE sub_category = $1
		ORDER BY id ASC`

	rows, err := DB.Query(query, subCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription, &part.SuperSSNumber,
			&part.Weight, &part.HsCode, &part.RemainPartNumber, &part.Coo, &part.RefNO,
			&part.Image, &part.MainCategory, &part.SubCategory,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	// Check for any iteration errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}

// PostEnquiry inserts a new enquiry into the database
func PostEnquiry(name, email, phone, enquiry, attachments string) (uint, error) {
	var id uint

	// Use current time as the created date
	currentTime := time.Now()

	// Execute the database query to insert the enquiry
	err := DB.QueryRow(
		"INSERT INTO enquiries (name, email, phone, enquiry, attachment, created_date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		name, email, phone, enquiry, attachments, currentTime,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	fmt.Println("Enquiry posted successfully")

	return id, nil
}

func GetEnquiries() ([]models.EnquiresModel, error) {
	rows, err := DB.Query(`
		SELECT id, name, email, phone, enquiry, attachment, created_date 
		FROM enquiries
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enquiriesList []models.EnquiresModel
	for rows.Next() {
		var enquiry models.EnquiresModel
		err := rows.Scan(
			&enquiry.ID,
			&enquiry.Name,
			&enquiry.Email,
			&enquiry.Phone,
			&enquiry.Enquiry,
			&enquiry.Attachments,
			&enquiry.CreatedDate,
		)
		if err != nil {
			return nil, err
		}
		enquiriesList = append(enquiriesList, enquiry)
	}

	fmt.Println("Get Enquiries Successful")

	return enquiriesList, nil
}

func GetPartByID(id string) (*models.MorrisParts, error) {
	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, images, main_category, sub_category,
		       dimension, compatible_engine_models, available_location, price
		FROM morrisparts
		WHERE id = $1
	`

	row := DB.QueryRow(query, id)

	var part models.MorrisParts
	var images sql.NullString // store JSON or comma-separated images

	err := row.Scan(
		&part.ID, &part.Name, &part.PartNumber, &part.PartDescription,
		&part.SuperSSNumber, &part.Weight, &part.HsCode, &part.RemainPartNumber,
		&part.Coo, &part.RefNO, &part.Image, &images,
		&part.MainCategory, &part.SubCategory, &part.Dimension,
		&part.CompatibleEngineModels, &part.AvailableLocation, &part.Price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("part not found")
		}
		return nil, err
	}

	// Convert images string to []string
	if images.Valid && images.String != "" {
		// if stored as JSON array
		if err := json.Unmarshal([]byte(images.String), &part.Images); err != nil {
			// fallback if comma-separated
			part.Images = strings.Split(images.String, ",")
		}
	}

	return &part, nil
}

func UpdateMorrisParts(
	id uint,
	name, partNumber, partDescription, superSSNumber, weight, hsCode, remainPartNumber,
	coo, refNo string,
	images []string, // Multiple images
	mainCategory, subCategory, dimension, compatibleEngineModels, availableLocation string,
	price float64,
) error {
	// Convert []string to JSON for storage
	imagesJSON, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	// Build query - cast to jsonb
	query := `
		UPDATE morrisparts 
		SET name = $1, 
		    part_number = $2, 
		    part_description = $3, 
		    super_ss_number = $4, 
		    weight = $5, 
		    hs_code = $6, 
		    remain_part_number = $7, 
		    coo = $8, 
		    ref_no = $9, 
		    images = $10::jsonb, 
		    main_category = $11, 
		    sub_category = $12,
		    dimension = $13,
		    compatible_engine_models = $14,
		    available_location = $15,
		    price = $16
		WHERE id = $17`

	// Execute update
	_, err = DB.Exec(query,
		name, partNumber, partDescription, superSSNumber, weight, hsCode,
		remainPartNumber, coo, refNo, string(imagesJSON),
		mainCategory, subCategory, dimension, compatibleEngineModels,
		availableLocation, price, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update morris parts: %w", err)
	}

	fmt.Println("✅ Update Morris Parts Successful")
	return nil
}

func GetRelatedParts(productID uint) ([]models.MorrisParts, error) {
	// First, get the main_category and sub_category of the given product
	var mainCategory, subCategory string
	err := DB.QueryRow(`
		SELECT main_category, sub_category 
		FROM morrisparts 
		WHERE id = $1
	`, productID).Scan(&mainCategory, &subCategory)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	// Now get 10 other products from the same category, excluding this product
	query := `
		SELECT id, name, part_number, part_description, super_ss_number, weight, hs_code,
		       remain_part_number, coo, ref_no, image, images, main_category, sub_category,
		       dimension, compatible_engine_models, available_location, price
		FROM morrisparts
		WHERE main_category = $1 AND sub_category = $2 AND id != $3
		ORDER BY id ASC
		LIMIT 10
	`

	rows, err := DB.Query(query, mainCategory, subCategory, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.MorrisParts
	for rows.Next() {
		var part models.MorrisParts
		var images pq.StringArray
		var dimension, compatibleEngineModels, availableLocation sql.NullString
		var price sql.NullFloat64

		err := rows.Scan(
			&part.ID, &part.Name, &part.PartNumber, &part.PartDescription, &part.SuperSSNumber,
			&part.Weight, &part.HsCode, &part.RemainPartNumber, &part.Coo, &part.RefNO,
			&part.Image, &images, &part.MainCategory, &part.SubCategory,
			&dimension, &compatibleEngineModels, &availableLocation, &price,
		)
		if err != nil {
			return nil, err
		}

		part.Images = []string(images)
		if dimension.Valid {
			part.Dimension = dimension.String
		}
		if compatibleEngineModels.Valid {
			part.CompatibleEngineModels = compatibleEngineModels.String
		}
		if availableLocation.Valid {
			part.AvailableLocation = availableLocation.String
		}
		if price.Valid {
			part.Price = price.Float64
		}

		parts = append(parts, part)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parts, nil
}
