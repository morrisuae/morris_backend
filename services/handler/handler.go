package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	"golang.org/x/exp/rand"
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
func generateRandomID() uint {
	// Seed the random number generator with uint64
	rand.Seed(uint64(time.Now().UnixNano()))

	// Generate a random number between 40000 and 90000
	randomID := rand.Uint32()%(90000-40000+1) + 40000

	return uint(randomID)
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
	ida := generateRandomID()
	id, err := helper.PostPart(ida, part.PartNumber, part.RemainPartNumber, part.PartDescription, part.FgWisonPartNumber, part.SuperSSNumber, part.Weight, part.Coo, part.HsCode, imageURL, part.SubCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

// PartCategory Handler
func PartCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostPartCategoryHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetPartCategoryHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostPartCategoryHandler(w http.ResponseWriter, r *http.Request) {

	var Category models.PartCategory

	if err := json.NewDecoder(r.Body).Decode(&Category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	id, err := helper.PostPartCategory(Category.ProductId, Category.ProductCategory, Category.Image, Category.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	Category.ID = id

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Category)
}

func GetPartCategoryHandler(w http.ResponseWriter, r *http.Request) {

	Category, err := helper.GetPartCategory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Category)

}

// PartCategory Handler
func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostCategoryHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetCategoryHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {
		DeleteCategoryHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
func PostCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var category models.Category

	// Get form values
	category.CategoryName = r.FormValue("category_name")

	// Process uploaded image
	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		fmt.Println("Error uploading file:", err)
		return
	}
	defer file.Close()

	// Read file bytes
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
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("CategoryImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
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

	// Construct imageURL
	imageURL := fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey)
	category.Image = imageURL

	// Save category to the database
	category.CreatedDate = time.Now()
	id, err := helper.PostCategory(category.Image, category.CategoryName, category.CreatedDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	category.ID = id

	// Respond with the created category
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func GetCategoryHandler(w http.ResponseWriter, r *http.Request) {

	category, err := helper.GetCategory()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)

}

func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = helper.DeleteCategory(uint(id))
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

// SubCategory Handler

func MorrisPartsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostMorrisPartsHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetPartsByCategoryHandler(w, r)
	} else if r.Method == http.MethodPut {
		UpdateMorrisParts(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteMorrisPartHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func MorrisPartsSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//	PostMorrisPartsHandler(w, r)
	} else if r.Method == http.MethodGet {
		SearchPartsHandler(w, r)
	} else if r.Method == http.MethodPut {
		// PutSubCategoryHandler(w, r)
	} else if r.Method == http.MethodDelete {
		//	DeleteMorrisPartHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetMorrisPartsHandler(w http.ResponseWriter, r *http.Request) {

	subCategory, err := helper.GetMorrisParts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subCategory)

}

func PostMorrisPartsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20) // 20MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var morrisPart models.MorrisParts

	// Parse form values
	morrisPart.Name = r.FormValue("name")
	morrisPart.PartNumber = r.FormValue("part_number")
	morrisPart.PartDescription = r.FormValue("part_description")
	morrisPart.SuperSSNumber = r.FormValue("super_ss_number")
	morrisPart.Weight = r.FormValue("weight")
	morrisPart.HsCode = r.FormValue("hs_code")
	morrisPart.RemainPartNumber = r.FormValue("remain_part_number")
	morrisPart.Coo = r.FormValue("coo")
	morrisPart.RefNO = r.FormValue("ref_no")
	morrisPart.MainCategory = r.FormValue("main_category")
	morrisPart.SubCategory = r.FormValue("sub_category")
	morrisPart.Dimension = r.FormValue("dimension")
	morrisPart.CompatibleEngineModels = r.FormValue("compatible_engine_models")
	morrisPart.AvailableLocation = r.FormValue("available_location")

	// Convert price
	if priceVal := r.FormValue("price"); priceVal != "" {
		if p, err := strconv.ParseFloat(priceVal, 64); err == nil {
			morrisPart.Price = p
		} else {
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
	}

	// Upload image to AWS S3
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"), // Replace with your AWS region
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	svc := s3.New(sess)

	// Helper function to upload image to S3
	uploadToS3 := func(fileBytes []byte, key string) (string, error) {
		if len(fileBytes) > 3*1024*1024 {
			img, _, _ := image.Decode(bytes.NewReader(fileBytes))
			newImage := resize.Resize(800, 0, img, resize.Lanczos3)
			var buf bytes.Buffer
			jpeg.Encode(&buf, newImage, nil)
			fileBytes = buf.Bytes()
		}
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String("morriuae"),
			Key:    aws.String(key),
			Body:   bytes.NewReader(fileBytes),
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", key), nil
	}

	// Upload main image
	if file, _, err := r.FormFile("image"); err == nil {
		fileBytes, _ := io.ReadAll(file)
		file.Close()

		imageKey := fmt.Sprintf("MorrisPartsImages/%d_main.jpg", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, imageKey)
		if err != nil {
			log.Printf("Failed to upload main image: %v", err)
			http.Error(w, "Main image upload failed", http.StatusInternalServerError)
			return
		}
		morrisPart.Image = uploadedURL
	}

	// Upload multiple images
	files := r.MultipartForm.File["images[]"]
	if len(files) == 0 {
		files = r.MultipartForm.File["images"] // fallback
	}

	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			continue
		}
		fileBytes, _ := io.ReadAll(src)
		src.Close()

		imageKey := fmt.Sprintf("MorrisPartsImages/%d_%s", time.Now().UnixNano(), f.Filename)
		uploadedURL, err := uploadToS3(fileBytes, imageKey)
		if err != nil {
			log.Printf("Failed to upload image %s: %v", f.Filename, err)
			continue
		}
		morrisPart.Images = append(morrisPart.Images, uploadedURL)
	}

	// Ensure images array is not null in JSON
	if morrisPart.Images == nil {
		morrisPart.Images = []string{}
	}

	// Save to DB
	id, err := helper.PostMorrisParts(
		morrisPart.Name,
		morrisPart.PartNumber,
		morrisPart.PartDescription,
		morrisPart.SuperSSNumber,
		morrisPart.Weight,
		morrisPart.HsCode,
		morrisPart.RemainPartNumber,
		morrisPart.Coo,
		morrisPart.RefNO,
		morrisPart.Image,
		morrisPart.MainCategory,
		morrisPart.SubCategory,
		morrisPart.Dimension,
		morrisPart.CompatibleEngineModels,
		morrisPart.AvailableLocation,
		morrisPart.Price,
		morrisPart.Images,
		helper.DB, // <-- Pass your DB instance here
	)
	if err != nil {
		fmt.Println("Error inserting Morris Part:", err)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	morrisPart.ID = id

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(morrisPart)
}

func DeleteMorrisPartHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the part ID from query parameters
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Call the helper function to delete the part
	err = helper.DeleteMorrisPart(uint(id))
	if err != nil {
		if err.Error() == "Part not found" {
			http.Error(w, "Part not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete part: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func BannersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostBannerHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetBannerHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {
		DeleteBannerHandler(w, r)
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

	var banner models.Banner

	// Parse form values
	banner.Title = r.FormValue("title")

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
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("SubCategoryImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
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
	banner.Image = imageURL

	// Save banner details into database
	id, err := helper.CreateBanner(banner.Image, banner.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	banner.ID = id
	banner.CreatedDate = time.Now()

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banner)
}

func GetBannerHandler(w http.ResponseWriter, r *http.Request) {
	banners, err := helper.GetBanners()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}
func DeleteBannerHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the banner ID from query parameters
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Call the helper function to delete the banner
	err = helper.DeleteBanner(uint(id))
	if err != nil {
		if err.Error() == "Banner not found" {
			http.Error(w, "Banner not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete banner: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HomeSliderBannerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostHomeSlideBannerHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetHomeSlideBannerHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {
		DeleteHomeSlideBannerHandler(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostHomeSlideBannerHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with a max size of 10MB
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var homeSlide models.HomeCompanySlides

	// Parse form values
	homeSlide.Name = r.FormValue("name")

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
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("HomeSlideImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
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
	homeSlide.Image = imageURL

	// Save home slide banner details into database
	id, err := helper.InsertHomeCompanySlide(homeSlide.Name, homeSlide.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	homeSlide.ID = id
	homeSlide.CreatedDate = time.Now()

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(homeSlide)
}
func GetHomeSlideBannerHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve home slide banners from the database
	homeSlides, err := helper.GetHomeCompanySlides()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(homeSlides)
}

func DeleteHomeSlideBannerHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the banner ID from query parameters
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Call the helper function to delete the home slide banner
	err = helper.DeleteHomeCompanySlide(uint(id))
	if err != nil {
		if err.Error() == "Home slide banner not found" {
			http.Error(w, "Home slide banner not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete home slide banner: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func SubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostSubCategoryHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetSubCategoriesByCategoryNameAndTypeHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostSubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with a max size of 10MB
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var subCategory models.SubCategory

	// Parse form values
	subCategory.MainCategoryName = r.FormValue("category_name")
	subCategory.SubCategoryName = r.FormValue("Sub_category_name")
	subCategory.CategoryType = r.FormValue("category_type")

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
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	imageKey := fmt.Sprintf("SubCategoryImage/%d.jpg", time.Now().Unix()) // Adjust key as needed
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
	subCategory.Image = imageURL

	// Save subcategory details into database
	id, err := helper.InsertSubCategory(subCategory.MainCategoryName, subCategory.SubCategoryName, subCategory.Image, subCategory.CategoryType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subCategory.ID = id
	subCategory.CreatedDate = time.Now()

	// Return response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subCategory)
}
func GetSubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve subcategories from the database
	subCategories, err := helper.GetSubCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(subCategories)
}

func GetSubCategoriesByCategoryNameAndTypeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the category name and category type from the request query parameters
	categoryName := r.URL.Query().Get("category_name")
	categoryType := r.URL.Query().Get("category_type")

	// Ensure that both categoryName and categoryType are provided
	if categoryName == "" || categoryType == "" {
		http.Error(w, "Both category_name and category_type are required", http.StatusBadRequest)
		return
	}

	// Call the helper function to get the subcategories
	subCategories, err := helper.GetSubCategoriesByCategoryNameAndType(categoryName, categoryType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(subCategories)
}

func GetPartsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
	mainCategory := r.URL.Query().Get("main_category")

	if mainCategory == "" {
		http.Error(w, "main_category is required", http.StatusBadRequest)
		return
	}

	parts, err := helper.GetPartsByCategory(mainCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

func SearchPartsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter
	partNumber := r.URL.Query().Get("part_number")

	// Validate query parameter
	if partNumber == "" {
		http.Error(w, "part_number is required", http.StatusBadRequest)
		return
	}

	// Fetch data using the helper function
	parts, err := helper.SearchParts(partNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

//ADMIN SIDE

func AdminPartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//	PostMorrisPartsHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetMorrisAdminPartsHandler(w, r)
	} else if r.Method == http.MethodPut {
		// PutSubCategoryHandler(w, r)
	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetMorrisAdminPartsHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve Morris parts from the database
	morrisParts, err := helper.GetMorrisParts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(morrisParts)
}

func AdminSubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//	PostMorrisPartsHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetAdminSubCategoriesHandler(w, r)
	} else if r.Method == http.MethodPut {
		// PutSubCategoryHandler(w, r)
	} else if r.Method == http.MethodDelete {
		DeleteAdminSubCategoryHandler(w, r)

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetAdminSubCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch subcategories using the helper function
	subCategories, err := helper.GetSubCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(subCategories)
}

func DeleteAdminSubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get the ID from query parameters
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Call the helper function to delete the subcategory
	err = helper.DeleteSubCategory(uint(id))
	if err != nil {
		if err.Error() == "Subcategory not found" {
			http.Error(w, "Subcategory not found", http.StatusNotFound)
			return
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete subcategory: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Respond with no content
	w.WriteHeader(http.StatusNoContent)
}

func UpdatePartHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT or PATCH
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the ID from the query parameter
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing ID parameter", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	// Parse the request body to extract the part details
	var part models.MorrisParts
	err = json.NewDecoder(r.Body).Decode(&part)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process the uploaded image if any
	file, _, err := r.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		fmt.Println("Error uploading file:", err)
		return
	}
	defer file.Close()

	var fileBytes []byte
	if file != nil {
		// Read file content
		fileBytes, err = ioutil.ReadAll(file)
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

			// Resize the image (set width to 800px and height auto)
			newImage := resize.Resize(800, 0, img, resize.Lanczos3)

			var buf bytes.Buffer
			err = jpeg.Encode(&buf, newImage, nil)
			if err != nil {
				http.Error(w, "Error encoding resized image", http.StatusInternalServerError)
				fmt.Println("Error encoding resized image:", err)
				return
			}
			fileBytes = buf.Bytes()
		}
	}

	// Upload image to AWS S3 if the file exists
	if len(fileBytes) > 0 {
		// Initialize the AWS session
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("eu-north-1"), // Replace with your AWS region
			Credentials: credentials.NewStaticCredentials(
				"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
				"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
				"",
			),
		})
		if err != nil {
			log.Printf("Failed to create AWS session: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		svc := s3.New(sess)
		imageKey := fmt.Sprintf("PartImages/%d.jpg", time.Now().Unix()) // Use a unique key for the image
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

		// Construct image URL
		part.Image = fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey)
	}

	// Set the ID of the part
	part.ID = uint(id)

	// Call the helper function to update the part
	err = helper.UpdatePart(part)
	if err != nil {
		if err.Error() == "part not found" {
			http.Error(w, "Part not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update part: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Respond with no content
	w.WriteHeader(http.StatusNoContent)
}

func GetPartsByOnlyCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//	PostMorrisPartsHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetPartsByOnlyCategory(w, r)
	} else if r.Method == http.MethodPut {
		// PutSubCategoryHandler(w, r)
	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetPartsByOnlyCategory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter
	mainCategory := r.URL.Query().Get("main_category")

	// Validate query parameter
	if mainCategory == "" {
		http.Error(w, "main_category is required", http.StatusBadRequest)
		return
	}

	// Fetch data using the helper function
	parts, err := helper.GetPartsOnlyByCategory(mainCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

func GetHomeSubCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {
		GetHomeSubCategories(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetHomeSubCategories(w http.ResponseWriter, r *http.Request) {
	// Call the helper function to get the subcategories
	subCategories, err := helper.GetSubCategoriesWithoutQuary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode and send the response
	json.NewEncoder(w).Encode(subCategories)
}

func GetPartsByOnlySubCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {
		GetPartsByOnlySubCategory(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func GetPartsByOnlySubCategory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter
	subCategory := r.URL.Query().Get("sub_category")

	// Validate query parameter
	if subCategory == "" {
		http.Error(w, "sub_category is required", http.StatusBadRequest)
		return
	}

	// Fetch data using the helper function
	parts, err := helper.GetPartsOnlyBySubCategory(subCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

///-------------------------------/////

func EnquiryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		HandlePostEnquiry(w, r)
	} else if r.Method == http.MethodGet {
		GetEnquiries(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func HandlePostEnquiry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a 10 MB limit
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract form fields
	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	enquiry := r.FormValue("enquiry")

	// Validate required fields
	if name == "" || email == "" || phone == "" || enquiry == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Extract file
	file, header, err := r.FormFile("attachment")
	if err != nil {
		http.Error(w, "Failed to read attachment", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file bytes
	fileBytes := new(bytes.Buffer)
	if _, err := fileBytes.ReadFrom(file); err != nil {
		http.Error(w, "Failed to read file bytes", http.StatusInternalServerError)
		return
	}

	// Upload file to S3
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",                     // Replace with your AWS access key ID
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S", // Replace with your AWS secret access key
			""), // Optional token, leave blank if not using
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	fileKey := fmt.Sprintf("attachments/%d_%s", time.Now().Unix(), header.Filename)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("morriuae"), // Replace with your S3 bucket name
		Key:    aws.String(fileKey),
		Body:   bytes.NewReader(fileBytes.Bytes()),
	})
	if err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Construct file URL
	attachmentURL := fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", fileKey)

	// Save enquiry to the database
	id, err := helper.PostEnquiry(name, email, phone, enquiry, attachmentURL)
	if err != nil {
		http.Error(w, "Failed to save enquiry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response object
	response := models.EnquiresModel{
		ID:          id,
		Name:        name,
		Email:       email,
		Phone:       phone,
		Enquiry:     enquiry,
		Attachments: attachmentURL,
		CreatedDate: time.Now(),
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func GetEnquiries(w http.ResponseWriter, r *http.Request) {
	enquiries, err := helper.GetEnquiries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enquiries)
}

func UpdateMorrisParts(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var morrisPart models.MorrisParts

	// Parse ID
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}
	morrisPart.ID = uint(id)

	// Fetch existing part to preserve images if needed
	existingPart, err := helper.GetPartByID(strconv.FormatUint(id, 10))
	if err != nil {
		http.Error(w, "Part not found", http.StatusNotFound)
		return
	}
	morrisPart.Images = existingPart.Images
	morrisPart.Image = existingPart.Image
	morrisPart.Price = existingPart.Price

	// Text fields
	morrisPart.Name = r.FormValue("name")
	morrisPart.PartNumber = r.FormValue("part_number")
	morrisPart.PartDescription = r.FormValue("part_description")
	morrisPart.SuperSSNumber = r.FormValue("super_ss_number")
	morrisPart.Weight = r.FormValue("weight")
	morrisPart.HsCode = r.FormValue("hs_code")
	morrisPart.RemainPartNumber = r.FormValue("remain_part_number")
	morrisPart.Coo = r.FormValue("coo")
	morrisPart.RefNO = r.FormValue("ref_no")
	morrisPart.MainCategory = r.FormValue("main_category")
	morrisPart.SubCategory = r.FormValue("sub_category")
	morrisPart.Dimension = r.FormValue("dimension")
	morrisPart.CompatibleEngineModels = r.FormValue("compatible_engine_models")
	morrisPart.AvailableLocation = r.FormValue("available_location")

	// Price
	if priceStr := r.FormValue("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			morrisPart.Price = price
		}
	}

	// AWS S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S",
			"",
		),
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	svc := s3.New(sess)

	// Handle multiple images
	files := r.MultipartForm.File["images"]
	if len(files) > 0 {
		var newImageURLs []string
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Failed to open file: %v", err)
				continue
			}
			defer file.Close()

			fileBytes, err := ioutil.ReadAll(file)
			if err != nil {
				log.Printf("Error reading file: %v", err)
				continue
			}

			// Resize if >3MB
			if len(fileBytes) > 3*1024*1024 {
				img, _, err := image.Decode(bytes.NewReader(fileBytes))
				if err == nil {
					newImg := resize.Resize(800, 0, img, resize.Lanczos3)
					var buf bytes.Buffer
					if jpeg.Encode(&buf, newImg, nil) == nil {
						fileBytes = buf.Bytes()
					}
				}
			}

			imageKey := fmt.Sprintf("MorrisPartsImages/%d_%s", time.Now().UnixNano(), fileHeader.Filename)
			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket: aws.String("morriuae"),
				Key:    aws.String(imageKey),
				Body:   bytes.NewReader(fileBytes),
			})
			if err != nil {
				log.Printf("Failed to upload to S3: %v", err)
				continue
			}

			newImageURLs = append(newImageURLs, fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey))
		}

		if len(newImageURLs) > 0 {
			morrisPart.Images = newImageURLs
		}
	}

	// Handle single image
	singleFile, singleHeader, err := r.FormFile("image")
	if err == nil {
		defer singleFile.Close()
		fileBytes, err := ioutil.ReadAll(singleFile)
		if err == nil {
			if len(fileBytes) > 3*1024*1024 {
				img, _, err := image.Decode(bytes.NewReader(fileBytes))
				if err == nil {
					newImg := resize.Resize(800, 0, img, resize.Lanczos3)
					var buf bytes.Buffer
					if jpeg.Encode(&buf, newImg, nil) == nil {
						fileBytes = buf.Bytes()
					}
				}
			}

			imageKey := fmt.Sprintf("MorrisPartsImages/%d_%s", time.Now().UnixNano(), singleHeader.Filename)
			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket: aws.String("morriuae"),
				Key:    aws.String(imageKey),
				Body:   bytes.NewReader(fileBytes),
			})
			if err == nil {
				morrisPart.Image = fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey)
			}
		}
	}

	// Update DB
	err = helper.UpdateMorrisParts(
		morrisPart.ID,
		morrisPart.Name,
		morrisPart.PartNumber,
		morrisPart.PartDescription,
		morrisPart.SuperSSNumber,
		morrisPart.Weight,
		morrisPart.HsCode,
		morrisPart.RemainPartNumber,
		morrisPart.Coo,
		morrisPart.RefNO,
		morrisPart.Image,  // single image
		morrisPart.Images, // multiple images
		morrisPart.MainCategory,
		morrisPart.SubCategory,
		morrisPart.Dimension,
		morrisPart.CompatibleEngineModels,
		morrisPart.AvailableLocation,
		morrisPart.Price,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Update successful"})
}

func PartByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		GetPartByID(w, r)
	} else if r.Method == http.MethodGet {
		GetPartByID(w, r)
	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
func GetPartByID(w http.ResponseWriter, r *http.Request) {
	// Get "id" from query params
	id := r.URL.Query().Get("id")

	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	// Call helper
	part, err := helper.GetPartByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(part)
}
func GetRelatedParts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {
		GetRelatedPartsHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
func GetRelatedPartsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse product ID from query
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		http.Error(w, "product_id is required", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product_id", http.StatusBadRequest)
		return
	}

	// Fetch related products
	parts, err := helper.GetRelatedParts(uint(productID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parts)
}

func GetEngines(w http.ResponseWriter, r *http.Request) {
	mainCategory := r.URL.Query().Get("main_category")

	var engines []models.Engine
	var err error

	if mainCategory != "" {
		// ✅ Fetch only engines with this main_category
		engines, err = helper.GetEnginesByMainCategory(mainCategory)
	} else {
		// ✅ Fetch all engines
		engines, err = helper.GetEngines()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engines)
}

func PostEngine(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var engine models.Engine
	engine.Name = r.FormValue("name")
	engine.PartNumber = r.FormValue("part_number")
	engine.Hz = r.FormValue("hz")
	engine.EpOrInd = r.FormValue("ep_or_ind")
	engine.Weight = r.FormValue("weight")
	engine.Coo = r.FormValue("coo")
	engine.Description = r.FormValue("description")
	engine.AvailableLocation = r.FormValue("available_location")
	engine.KVA = r.FormValue("kva")
	engine.MainCategory = r.FormValue("main_category")
	engine.CreatedDate = time.Now()

	// AWS S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S",
			"",
		),
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, "AWS session error", http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	uploadToS3 := func(fileBytes []byte, key string) (string, error) {
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String("morriuae"),
			Key:         aws.String(key),
			Body:        bytes.NewReader(fileBytes),
			ContentType: aws.String(http.DetectContentType(fileBytes)),
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", key), nil
	}

	// Upload main image
	if file, _, err := r.FormFile("image"); err == nil {
		fileBytes, _ := io.ReadAll(file)
		file.Close()

		imageKey := fmt.Sprintf("EngineImages/%d_main.jpg", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, imageKey)
		if err != nil {
			http.Error(w, "Image upload failed", http.StatusInternalServerError)
			return
		}
		engine.Image = uploadedURL
	}

	// Upload multiple images
	files := r.MultipartForm.File["images[]"]
	if len(files) == 0 {
		files = r.MultipartForm.File["images"] // fallback
	}

	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			continue
		}
		fileBytes, _ := io.ReadAll(src)
		src.Close()

		imageKey := fmt.Sprintf("EngineImages/%d_%s", time.Now().UnixNano(), f.Filename)
		uploadedURL, err := uploadToS3(fileBytes, imageKey)
		if err != nil {
			log.Printf("Failed to upload image %s: %v", f.Filename, err)
			continue
		}
		engine.Images = append(engine.Images, uploadedURL)
	}

	if engine.Images == nil {
		engine.Images = []string{}
	}

	// Upload Specification PDF
	if file, header, err := r.FormFile("specification_pdf"); err == nil {
		defer file.Close()
		fileBytes, _ := io.ReadAll(file)

		if filepath.Ext(header.Filename) != ".pdf" {
			http.Error(w, "Only PDF files are allowed for specification", http.StatusBadRequest)
			return
		}

		pdfKey := fmt.Sprintf("EngineSpecifications/%d_spec.pdf", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, pdfKey)
		if err != nil {
			http.Error(w, "Specification PDF upload failed", http.StatusInternalServerError)
			return
		}
		engine.SpecificationURL = uploadedURL
	}

	// Save to DB
	id, err := helper.PostEngine(engine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	engine.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engine)
}

func EngineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostEngine(w, r)
	} else if r.Method == http.MethodGet {
		GetEngines(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostCatalogue(w http.ResponseWriter, r *http.Request) {
	// Limit multipart form size
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var catalogue models.Catalogue
	catalogue.Title = r.FormValue("title")
	catalogue.PartNumber = r.FormValue("part_number")
	catalogue.Description = r.FormValue("description")
	catalogue.MainCategory = r.FormValue("main_category")
	catalogue.CreatedDate = time.Now()

	// AWS S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S",
			"",
		),
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, "AWS session error", http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)
	uploadToS3 := func(fileBytes []byte, key string, contentType string) (string, error) {
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String("morriuae"),
			Key:         aws.String(key),
			Body:        bytes.NewReader(fileBytes),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", key), nil
	}

	// Upload image (optional)
	if file, _, err := r.FormFile("image"); err == nil {
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			http.Error(w, "Failed to read image file", http.StatusInternalServerError)
			return
		}

		imageKey := fmt.Sprintf("CatalogueImages/%d_main.jpg", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, imageKey, http.DetectContentType(fileBytes))
		if err != nil {
			http.Error(w, "Image upload failed", http.StatusInternalServerError)
			return
		}
		catalogue.Image = uploadedURL
	}

	// Upload PDF (optional)
	file, header, err := r.FormFile("pdf_url")
	if err != nil {
		log.Printf("No PDF uploaded or error reading file: %v", err)
	} else {
		defer file.Close()
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read PDF file", http.StatusInternalServerError)
			return
		}

		if strings.ToLower(filepath.Ext(header.Filename)) != ".pdf" {
			http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		pdfKey := fmt.Sprintf("CataloguePDFs/%d_file.pdf", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, pdfKey, "application/pdf")
		if err != nil {
			http.Error(w, "PDF upload failed", http.StatusInternalServerError)
			return
		}
		catalogue.PdfURL = uploadedURL
	}

	// Save to DB
	id, err := helper.PostCatalogue(catalogue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	catalogue.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogue)
}

func GetCatalogues(w http.ResponseWriter, r *http.Request) {
	mainCategory := r.URL.Query().Get("main_category")

	var catalogues []models.Catalogue
	var err error

	if mainCategory != "" {
		// Fetch only catalogues with this main_category
		catalogues, err = helper.GetCataloguesByMainCategory(mainCategory)
	} else {
		// Fetch all catalogues
		catalogues, err = helper.GetCatalogues()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogues)
}

func CatalogeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostCatalogue(w, r)
	} else if r.Method == http.MethodGet {
		GetCatalogues(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func EngineHandlerByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {
		GetEnginesByID(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
func CatalogueHandlerByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else if r.Method == http.MethodGet {
		GetCatalogueByID(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
func GetEnginesByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	mainCategory := r.URL.Query().Get("main_category")

	if idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Invalid id", http.StatusBadRequest)
			return
		}
		engine, err := helper.GetEngineByID(uint(id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if engine == nil {
			http.Error(w, "Engine not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(engine)
		return
	}

	var engines []models.Engine
	var err error
	if mainCategory != "" {
		engines, err = helper.GetEnginesByMainCategory(mainCategory)
	} else {
		engines, err = helper.GetEngines()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(engines)
}

func GetCatalogueByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	if idParam == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	catalogue, err := helper.GetCatalogueByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if catalogue == nil {
		http.Error(w, "Catalogue not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalogue)
}

func CustomerDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostCustomerDetails(w, r)
	} else if r.Method == http.MethodGet {
		GetCustomerDetailsHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}

func PostCustomerDetails(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var customer models.CustomerDetails
	customer.Name = r.FormValue("name")
	customer.Phone = r.FormValue("phone")
	customer.Email = r.FormValue("email")
	customer.CompanyName = r.FormValue("company_name")
	customer.Country = r.FormValue("country")
	customer.CreatedDate = time.Now()

	// Save to DB
	id, err := helper.PostCustomerDetails(customer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	customer.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func GetCustomerDetailsHandler(w http.ResponseWriter, r *http.Request) {
	customers, err := helper.GetCustomerDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}

func GetBrandCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	brandCategories, err := helper.GetBrandCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brandCategories)
}

func PostBrandCategory(w http.ResponseWriter, r *http.Request) {
	// Limit multipart form size (20 MB)
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var brandCategory models.BrandCategory
	brandCategory.Name = r.FormValue("name")
	brandCategory.MainCategory = r.FormValue("main_category") // 👈 New field
	brandCategory.CreatedDate = time.Now()

	// AWS S3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
		Credentials: credentials.NewStaticCredentials(
			"AKIAWMFUPPBUFJOAZMAT",
			"kFHNm5UvPvBcEDiFi6p3sRuej9oruy6kSYkkjk/S",
			"",
		),
	})
	if err != nil {
		log.Printf("Failed to create AWS session: %v", err)
		http.Error(w, "AWS session error", http.StatusInternalServerError)
		return
	}

	svc := s3.New(sess)

	uploadToS3 := func(fileBytes []byte, key string, contentType string) (string, error) {
		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String("morriuae"),
			Key:         aws.String(key),
			Body:        bytes.NewReader(fileBytes),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", key), nil
	}

	// Upload image (optional)
	if file, _, err := r.FormFile("image"); err == nil {
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			http.Error(w, "Failed to read image file", http.StatusInternalServerError)
			return
		}

		imageKey := fmt.Sprintf("BrandCategory/%d.jpg", time.Now().Unix())
		uploadedURL, err := uploadToS3(fileBytes, imageKey, http.DetectContentType(fileBytes))
		if err != nil {
			http.Error(w, "Image upload failed", http.StatusInternalServerError)
			return
		}
		brandCategory.Image = uploadedURL
	}

	// Save to DB
	id, err := helper.PostBrandCategory(brandCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brandCategory.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brandCategory)
}

func BrandCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostBrandCategory(w, r)
	} else if r.Method == http.MethodGet {
		GetBrandCategoriesHandler(w, r)
	} else if r.Method == http.MethodPut {

	} else if r.Method == http.MethodDelete {

	} else {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
	}
}
