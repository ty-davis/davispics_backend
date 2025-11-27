package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"
)

var s3Client *s3.Client

func InitS3Client() error {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               os.Getenv("R2_ENDPOINT"),
			SigningRegion:     "auto",
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return err
	}

	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return nil
}

func UploadImages(r *http.Request) ([]string, error) {
	// Parse multipart form (32MB max memory)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, fmt.Errorf("failed to parse form data: %w", err)
	}

	// Get form values
	password := r.FormValue("password")
	maxDimensionStr := r.FormValue("maxDimension")
	folder := r.FormValue("folder")

	// validate password
	expectedPassword := os.Getenv("UPLOAD_PASSWORD")
	if expectedPassword == "" {
		return nil, fmt.Errorf("Server config error: No upload password set")
	}
	if password != expectedPassword {
		return nil, fmt.Errorf("Invalid password")
	}

	maxDimension := 0
	if maxDimensionStr != "" {
		maxDimension, err = strconv.Atoi(maxDimensionStr)
		if err != nil || maxDimension <= 0 {
			return nil, fmt.Errorf("Invalid maxDimension value")
		}
	}

	if folder == "" {
		return nil, fmt.Errorf("Folder name is required")
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 {
		return nil, fmt.Errorf("No files provided")
	}

	log.Printf("Received %d files for folder '%s' with maxDimension %d\n", len(files), folder, maxDimension)

	var uploadedURLs []string
	for _, fileHeader := range files {
		// Validate file type
		if !isValidImageType(fileHeader) {
			return nil, fmt.Errorf("Invalid file type: %s (only PNG and JPEG allowed)", fileHeader.Filename)
		}

		// Upload to R2
		url, err := uploadToR2(fileHeader, folder, maxDimension)
		if err != nil {
			return nil, fmt.Errorf("Failed to upload %s: %w", fileHeader.Filename, err)
		}
		uploadedURLs = append(uploadedURLs, url)
	}

	return uploadedURLs, nil
}

func isValidImageType(fileHeader *multipart.FileHeader) bool {
	contentType := fileHeader.Header.Get("Content-Type")
	return contentType == "image/png" || contentType == "image/jpeg"
}

func uploadToR2(fileHeader *multipart.FileHeader, folder string, maxDimension int) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if maxDimension is specified
	if maxDimension > 0 {
		img = resizeImage(img, maxDimension)
	}

	// Encode image to buffer
	var buf bytes.Buffer
	var contentType string

	if format == "jpeg" || format == "jpg" {
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		contentType = "image/jpeg"
	} else if format == "png" {
		err = png.Encode(&buf, img)
		contentType = "image/png"
	} else {
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	bucket := os.Getenv("R2_BUCKET_NAME")
	key := filepath.Join(folder, generateUniqueFilename(fileHeader.Filename))

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", err
	}
	publicURL := fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_URL"), key)

	log.Printf("Uploaded %s to R2 as %s\n", fileHeader.Filename, key)
	return publicURL, nil
}

func resizeImage(img image.Image, maxDimension int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width <= maxDimension && height <= maxDimension {
		return img
	}

	var newWidth, newHeight uint
	if width > height {
		newWidth = uint(maxDimension)
		newHeight = 0 // maintain aspect ratio
	} else {
		newWidth = 0 // maintain aspect ratio
		newHeight = uint(maxDimension)
	}

	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	base := strings.TrimSuffix(originalFilename, ext)
	return fmt.Sprintf("%s_%d%s", base, time.Now().Unix(), ext)
}
