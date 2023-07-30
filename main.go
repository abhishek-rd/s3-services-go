package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"s3service"
)

func main() {
	router := gin.Default()

	s3Service, err := s3service.New()
	if err != nil {
		panic("Failed to initialize S3 Service: " + err.Error())
	}

	router.GET("/get-json", func(c *gin.Context) {
		data, err := s3Service.GetObject("yourBucketName", "yourFileName.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get object from S3"})
			return
		}

		c.JSON(http.StatusOK, data)
	})

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		err = s3Service.UploadObject("yourBucketName", file.Filename, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to upload file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	})

	router.Run(":8080")
}
