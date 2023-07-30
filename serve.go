package main

import (
	"github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"encoding/json"
)

func main() {
	router := gin.Default()

	router.GET("/get-json", func(c *gin.Context) {
		bucket := "stt-test-framework-dev"
		key := "suites.json"

		// Create a session
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String("us-west-2")},
		)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create session," + err.Error()})
			return
		}

		// Create a downloader with the session and default options
		downloader := s3manager.NewDownloader(sess)

		// Write the contents of S3 Object to a buffer
		buff := &aws.WriteAtBuffer{}
		_, err = downloader.Download(buff,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to download file," + err.Error()})
			return
		}

		// Convert bytes to a map
		var dat map[string]interface{}
		if err := json.Unmarshal(buff.Bytes(), &dat); err != nil {
			c.JSON(500, gin.H{"error": "Failed to parse JSON," + err.Error()})
			return
		}

		// Return the content of the JSON
		c.JSON(200, dat)
	})

	router.Run() // listen and serve on 0.0.0.0:8080
}
