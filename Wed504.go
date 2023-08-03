package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"mime/multipart"
)

type Test struct {
	File             string `json:"file"`
	GoldenTranscript string `json:"goldenTranscript"`
	STTResult        string `json:"sttResult,omitempty"`
}

type Tests struct {
	Tests []Test `json:"tests"`
}

func main() {
	r := gin.Default()
	r.POST("/upload", handleUpload)
	r.Run()
}

func handleUpload(c *gin.Context) {
	form, _ := c.MultipartForm()
	csvFile, _ := c.FormFile("csv")
	oggFiles := form.File["oggFiles"]

	tests := parseCSVToJSON(csvFile)

	for _, file := range oggFiles {
		openedFile, err := file.Open()
		if err != nil {
			log.Printf("error opening file: %v", err)
			continue
		}

		transcription := transcribeLocalFile(openedFile)

		for idx, test := range tests.Tests {
			if test.File == file.Filename {
				tests.Tests[idx].STTResult = transcription
			}
		}

		uploadToS3(file)

		if err := openedFile.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}

	jsonData, err := json.Marshal(tests)
	if err != nil {
		log.Fatal("Could not convert to JSON", err)
	}

	c.JSON(200, jsonData)
}

func parseCSVToJSON(file *multipart.FileHeader) Tests {
	f, _ := file.Open()
	defer f.Close()

	reader := csv.NewReader(f)
	lines, _ := reader.ReadAll()

	var tests Tests
	for _, line := range lines {
		tests.Tests = append(tests.Tests, Test{
			File:             line[0],
			GoldenTranscript: line[1],
		})
	}

	return tests
}

func uploadToS3(file *multipart.FileHeader) {
	bucket := "Your_BUCKET_NAME"

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
		Bucket: aws.String(bucket),
		Key:    aws.String(file.Filename),
		Body:   f,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func transcribeLocalFile(file io.Reader) string {
	sess := session.Must(session.NewSession())
	client := transcribestreamingservice.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	lc := "en-US"
	me := "ogg-opus"

	resp, err := client.StartStreamTranscription(&transcribestreamingservice.StartStreamTranscriptionInput{
		LanguageCode:         &lc,
		MediaEncoding:        &me,
		MediaSampleRateHertz: aws.Int64(16000),
	})
	if err != nil {
		log.Fatalf("failed to start streaming, %v", err)
	}

	stream := resp.GetStream()
	defer stream.Close()

	transcribestreamingservice.StreamAudioFromReader(context.Background(), stream.Writer, 10*1024, file)
	var textout string
	for event := range stream.Events() {
		switch e := event.(type) {
		case *transcribestreamingservice.TranscriptEvent:
			for _, res := range e.Transcript.Results {
				if !*res.IsPartial {
					for _, alt := range res.Alternatives {
						textout += aws.StringValue(alt.Transcript) + " "
					}
				}
			}
		default:
			log.Fatalf("unexpected event, %T", event)
		}
	}

	return textout
}
