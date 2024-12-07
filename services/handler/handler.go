package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	"morris-backend.com/main/services/helper"
	"morris-backend.com/main/services/models"
)

// Part GET, POST, PUT and DELETE
func PartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostPartHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetPartHandler(w, r)
	} else if r.Method == http.MethodPut {
		PutPartHandler(w, r)
	} else if r.Method == http.MethodDelete {
		DeletePartHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostPartHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var part models.Part

	part.PartNumber = r.FormValue("title")
	part.PartNumber = r.FormValue("part_number")
	part.RemainPartNumber = r.FormValue("remain_part_number")
	part.PartDescription = r.FormValue("part_description")
	part.FgWisonPartNumber = r.FormValue("fg_wison_part_number")
	part.SuperSSNumber = r.FormValue("super_ss_number")
	part.Weight = r.FormValue("weight")
	part.Coo = r.FormValue("coo")
	part.HsCode = r.FormValue("hs_code")
	part.SubCategory = r.FormValue("sub_category")

	// Process uploaded image
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		fmt.Println("Error uploading file:", err)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file content", http.StatusInternalServerError)
		fmt.Println("Error reading file content:", err)
		return
	}

	// Resize image if it exceeds 3MB
	if len(fileBytes) > 3*1024*1024 {
		img, _, err := image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			http.Error(w, "Error decoding image", http.StatusInternalServerError)
			fmt.Println("Error decoding image:", err)
			return
		}

		newImage := resize.Resize(800, 0, img, resize.Lanczos3)
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, newImage, nil)
		if err != nil {
			http.Error(w, "Error encoding compressed image", http.StatusInternalServerError)
			fmt.Println("Error encoding compressed image:", err)
			return
		}
		fileBytes = buf.Bytes()
	}

	// Upload image to AWS S3
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"), // Replace with your AWS region
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUKH747TDX",                     // Replace with your AWS access key ID
			"AvUBkX2gtAFWupNIBdCr9BtZUDbPtdd/Vj30Hj4J", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("PartImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("morriuae"), // Replace with your S3 bucket name
		Key:    aws.String(imageKey),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		log.Printf("Failed to upload image to S3: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Construct imageURL assuming it's from your S3 bucket
	imageURL := fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey)
	id, err := helper.PostPart(part.PartNumber, part.RemainPartNumber, part.PartDescription, part.FgWisonPartNumber, part.SuperSSNumber, part.Weight, part.Coo, part.HsCode, imageURL, part.SubCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	part.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(part)
}

func GetPartHandler(w http.ResponseWriter, r *http.Request) {

	part, err := helper.GetPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(part)

}

func GetPartHandlerByPartNumber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract part_number from URL query parameter
	partNumber := r.URL.Query().Get("part_number")
	if partNumber == "" {
		http.Error(w, "part_number parameter is required", http.StatusBadRequest)
		return
	}

	// Retrieve parts from repository
	parts, err := helper.GetPartByPartNumber(partNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(parts) == 0 {
		http.Error(w, "No parts found", http.StatusNotFound)
		return
	}

	// Serialize parts to JSON and write response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(parts)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func PutPartHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var part models.Part
	if err := json.NewDecoder(r.Body).Decode(&part); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the id from the query parameter
	part.ID = uint(id)

	err = helper.PutPart(part.ID, part.PartNumber, part.RemainPartNumber, part.PartDescription, part.FgWisonPartNumber, part.SuperSSNumber, part.Weight, part.Coo, part.HsCode, part.Image, part.SubCategory)
	if err != nil {
		if err.Error() == "Log not found" {
			http.Error(w, "Log not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, fmt.Sprintf("Failed to update log: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(part); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DeletePartHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = helper.DeletePart(uint(id))
	if err != nil {
		if err.Error() == "Part not found" {
			http.Error(w, "part not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete part: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Banner GET and POST
func BannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostBannerHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetBannerHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostBannerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var Banner models.Banner

	// Process uploaded image
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		fmt.Println("Error uploading file:", err)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file content", http.StatusInternalServerError)
		fmt.Println("Error reading file content:", err)
		return
	}

	// Resize image if it exceeds 3MB
	if len(fileBytes) > 3*1024*1024 {
		img, _, err := image.Decode(bytes.NewReader(fileBytes))
		if err != nil {
			http.Error(w, "Error decoding image", http.StatusInternalServerError)
			fmt.Println("Error decoding image:", err)
			return
		}

		newImage := resize.Resize(800, 0, img, resize.Lanczos3)
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, newImage, nil)
		if err != nil {
			http.Error(w, "Error encoding compressed image", http.StatusInternalServerError)
			fmt.Println("Error encoding compressed image:", err)
			return
		}
		fileBytes = buf.Bytes()
	}

	// Upload image to AWS S3
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your AWS region
		Credentials: credentials.NewStaticCredentials(
			"AKIAYS2NVN4MBSHP33FF",                     // Replace with your AWS access key ID
			"aILySGhiQAB7SaFnqozcRZe1MhZ0zNODLof2Alr4", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("BannerImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("tendercall-db"), // Replace with your S3 bucket name
		Key:    aws.String(imageKey),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		log.Printf("Failed to upload image to S3: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Construct imageURL assuming it's from your S3 bucket
	imageURL := fmt.Sprintf("https://tendercall-db.s3.amazonaws.com/%s", imageKey)

	err = helper.PostBanner(imageURL, Banner.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Banner)
}

func GetBannerHandler(w http.ResponseWriter, r *http.Request) {

	banner, err := helper.GetBanner()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banner)

}

// Company Handler
func CompanyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostCompanyHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetCompanyHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostCompanyHandler(w http.ResponseWriter, r *http.Request) {

	var company models.Company

	if err := json.NewDecoder(r.Body).Decode(&company); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	id, err := helper.PostCompany(company.CompanyName, company.CoverImage, company.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	company.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(company)
}

func GetCompanyHandler(w http.ResponseWriter, r *http.Request) {

	company, err := helper.GetCompany()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(company)

}

// func GetPartHandlerByPartNumber(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Extract part_number from URL query parameter
// 	partNumber := r.URL.Query().Get("part_number")
// 	if partNumber == "" {
// 		http.Error(w, "part_number parameter is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Retrieve parts from repository
// 	parts, err := helper.GetPartByPartNumber(partNumber)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	if len(parts) == 0 {
// 		http.Error(w, "No parts found", http.StatusNotFound)
// 		return
// 	}

// 	// Serialize parts to JSON and write response
// 	w.Header().Set("Content-Type", "application/json")
// 	err = json.NewEncoder(w).Encode(parts)
// 	if err != nil {
// 		http.Error(w, "Error encoding response", http.StatusInternalServerError)
// 		return
// 	}
// }

// Category Handler
func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostCategoryHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetCategoryHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostCategoryHandler(w http.ResponseWriter, r *http.Request) {

	var Category models.PartCategory

	if err := json.NewDecoder(r.Body).Decode(&Category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	id, err := helper.PostCategory(Category.ProductId, Category.ProductCategory, Category.Image, Category.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	Category.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Category)
}

func GetCategoryHandler(w http.ResponseWriter, r *http.Request) {

	Category, err := helper.GetCategory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Category)

}
