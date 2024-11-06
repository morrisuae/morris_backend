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
func PostPart(part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code string) (uint, error) {
	// Connect to the database
	var id uint

	err := DB.QueryRow("INSERT INTO parts (part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code).Scan(&id)
	if err != nil {
		return 0, err
	}

	fmt.Println("Post Successful")

	return id, nil
}

func GetPart() ([]models.Part, error) {
	rows, err := DB.Query("SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code FROM parts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []models.Part
	for rows.Next() {
		var part models.Part
		err := rows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode)
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
		fullPartQuery := "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code FROM parts WHERE part_number = $1"
		fullPartRows, err := DB.Query(fullPartQuery, partNumber)
		if err != nil {
			log.Println("Error executing full part query:", err)
			return nil, err
		}
		defer fullPartRows.Close()

		// Check for exact match
		for fullPartRows.Next() {
			var part models.Part
			err := fullPartRows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode)
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
		query = "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code FROM parts WHERE part_number LIKE $1"
		args = append(args, partNumber[:3]+"%")
	} else {
		// Handle as an exact match if partNumber is shorter than 3 characters
		query = "SELECT id, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code FROM parts WHERE part_number = $1"
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
		err := rows.Scan(&part.ID, &part.PartNumber, &part.RemainPartNumber, &part.PartDescription, &part.FgWisonPartNumber, &part.SuperSSNumber, &part.Weight, &part.Coo, &part.HsCode)
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

func PutPart(id uint, part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code string) error {
	result, err := DB.Exec("UPDATE parts SET part_number=$1, remain_part_number=$2, part_description=$3, fg_wison_part_number=$4, super_ss_number=$5, weight=$6, coo=$7, hs_code=$8 WHERE id=$9", part_number, remain_part_number, part_description, fg_wison_part_number, super_ss_number, weight, coo, hs_code, id)

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
