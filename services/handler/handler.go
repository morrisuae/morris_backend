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
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
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
	imageKey := fmt.Sprintf("MorrisPartsImages/%d.jpg", time.Now().Unix())
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("morriuae"),
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

	// Save part details into database
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
		imageURL,
		morrisPart.MainCategory,
		morrisPart.SubCategory,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	morrisPart.ID = id

	// Return response as JSON
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
	// Parse query parameters
	mainCategory := r.URL.Query().Get("main_category")
	subCategory := r.URL.Query().Get("sub_category")

	// Validate query parameters
	if mainCategory == "" || subCategory == "" {
		http.Error(w, "Both main_category and sub_category are required", http.StatusBadRequest)
		return
	}

	// Fetch data using the helper function
	parts, err := helper.GetPartsByCategory(mainCategory, subCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response as JSON
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
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var morrisPart models.MorrisParts

	// Parse form values
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid or missing ID", http.StatusBadRequest)
		return
	}
	morrisPart.ID = uint(id)

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

	// Process uploaded image
	file, _, err := r.FormFile("image")
	var imageURL string
	if err == nil { // Image provided
		defer file.Close()

		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file content", http.StatusInternalServerError)
			return
		}

		// Resize image if it exceeds 3MB
		if len(fileBytes) > 3*1024*1024 {
			img, _, err := image.Decode(bytes.NewReader(fileBytes))
			if err != nil {
				http.Error(w, "Error decoding image", http.StatusInternalServerError)
				return
			}

			newImage := resize.Resize(800, 0, img, resize.Lanczos3)
			var buf bytes.Buffer
			err = jpeg.Encode(&buf, newImage, nil)
			if err != nil {
				http.Error(w, "Error encoding compressed image", http.StatusInternalServerError)
				return
			}
			fileBytes = buf.Bytes()
		}

		// Upload image to AWS S3
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
		imageKey := fmt.Sprintf("MorrisPartsImages/%d.jpg", time.Now().Unix())
		_, err = svc.PutObject(&s3.PutObjectInput{
			Bucket: aws.String("morriuae"),
			Key:    aws.String(imageKey),
			Body:   bytes.NewReader(fileBytes),
		})
		if err != nil {
			log.Printf("Failed to upload image to S3: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Construct imageURL
		imageURL = fmt.Sprintf("https://morriuae.s3.amazonaws.com/%s", imageKey)
	}

	// Call helper function to update data
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
		imageURL,
		morrisPart.MainCategory,
		morrisPart.SubCategory,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response as JSON
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
