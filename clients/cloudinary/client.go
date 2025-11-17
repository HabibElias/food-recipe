package cloudinary

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

func Init() {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		log.Println("Cloudinary credentials not set; Cloudinary client will not be initialized")
		return
	}

	var err error
	cld, err = cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Printf("Failed to initialize Cloudinary client: %v", err)
		return
	}

	log.Println("cloudinary: client initialized")
}

func Get() *cloudinary.Cloudinary {
	return cld
}

func UploadBase64Image(ctx context.Context, base64String, publicID, folder string) (string, error) {
	if cld == nil {
		return "", fmt.Errorf("cloudinary client not initialized")
	}

	result, err := cld.Upload.Upload(ctx, base64String, uploader.UploadParams{
		PublicID:     publicID,
		Folder:       folder,
		ResourceType: "image",
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	return result.SecureURL, nil
}
