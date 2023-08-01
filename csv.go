// main.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
)

type Test struct {
	File             string `json:"file"`
	GoldenTranscript string `json:"goldenTranscript"`
}

type Tests struct {
	Tests []Test `json:"tests"`
}

func main() {
	r := gin.Default()

	r.POST("/upload", func(c *gin.Context) {
		csvFile, _ := c.FormFile("csv")
		oggFiles := c.MultipartForm.File["oggFiles"]

		uploadToS3(csvFile)

		for _, file := range oggFiles {
			uploadToS3(file)
		}

		c.JSON(200, gin.H{
			"message": "Files uploaded to S3",
		})
	})

	r.Run()
}

func uploadToS3(file *multipart.FileHeader) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String("REGION"),
		Credentials: credentials.NewStaticCredentials("Your_AWS_ACCESS_KEY", "Your_AWS_SECRET_KEY", ""),
	})

	if err != nil {
		log.Fatal(err)
	}

	uploader := s3manager.NewUploader(s)

	f, _ := file.Open()
	defer f.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("Your_BUCKET_NAME"),
		Key:    aws.String(file.Filename),
		Body:   io.Reader(f),
	})

	if err != nil {
		log.Fatal(err)
	}

	if file.Header.Get("Content-Type") == "text/csv" {
		parseCSVToJSON(f, file.Filename+".json")
	}
}

func parseCSVToJSON(csvFile io.Reader, jsonFileName string) {
	reader := csv.NewReader(csvFile)

	lines, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Could not read the CSV file", err)
	}

	var tests Tests
	for _, line := range lines {
		tests.Tests = append(tests.Tests, Test{
			File:             line[0],
			GoldenTranscript: line[1],
		})
	}

	jsonData, err := json.Marshal(tests)
	if err != nil {
		log.Fatal("Could not convert to JSON", err)
	}

	err = ioutil.WriteFile(jsonFileName, jsonData, 0644)
	if err != nil {
		log.Fatal("Could not write the JSON file", err)
	}
}
