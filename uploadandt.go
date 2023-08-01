// main.go
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
)

type Test struct {
	File             string `json:"file"`
	GoldenTranscript string `json:"goldenTranscript"`
	STTResult        string `json:"sttResult,omitempty"`
}

type Tests struct {
	Tests []Test `json:"tests"`
}

var tests Tests
var bucket = "Your_BUCKET_NAME"  // replace with your bucket name

func main() {
	r := gin.Default()

	r.POST("/upload", func(c *gin.Context) {
		form, _ := c.MultipartForm()

		csvFile, _ := c.FormFile("csv")
		oggFiles := form.File["oggFiles"]

		uploadAndTranscribe(c, csvFile, oggFiles)

		c.JSON(200, gin.H{
			"message": "Files uploaded to S3 and transcribed",
		})
	})

	r.Run()
}

func uploadAndTranscribe(c *gin.Context, csvFile *multipart.FileHeader, oggFiles []*multipart.FileHeader) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String("REGION"),
	})

	if err != nil {
		log.Fatal(err)
	}

	uploader := s3manager.NewUploader(s)

	// Upload and parse CSV
	src, _ := csvFile.Open()
	defer src.Close()

	reader, _ := csv.NewReader(src).ReadAll()
	parseCSVToMemory(reader)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(csvFile.Filename),
		Body:   src,
	})

	if err != nil {
		log.Fatal(err)
	}

	// Upload and transcribe OGG files
	for _, file := range oggFiles {
		oggSrc, _ := file.Open()
		defer oggSrc.Close()

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(file.Filename),
			Body:   oggSrc,
		})

		if err != nil {
			log.Fatal(err)
		}

		transcription := transcribeAudio(file.Filename)
		updateTestsInMemory(file.Filename, transcription)
	}

	saveTestsToFile("output.json")
}

func parseCSVToMemory(lines [][]string) {
	tests = Tests{}  // Reset tests
	for _, line := range lines {
		tests.Tests = append(tests.Tests, Test{
			File:             line[0],
			GoldenTranscript: line[1],
		})
	}
}

func updateTestsInMemory(filename string, transcription string) {
	// update STTResult for the matching file
	for i, test := range tests.Tests {
		if test.File == filename {
			tests.Tests[i].STTResult = transcription
		}
	}
}

func saveTestsToFile(jsonFileName string) {
	// save JSON file
	jsonData, err := json.Marshal(tests)
	if err != nil {
		log.Fatal("Could not convert to JSON", err)
	}
	err = ioutil.WriteFile(jsonFileName, jsonData, 0644)
	if err != nil {
		log.Fatal("Could not write the JSON file", err)
	}
}

func transcribeAudio(filename string) string {
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

	audioFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open audio file, %v", err)
	}
	defer audioFile.Close()

	transcribestreamingservice.StreamAudioFromReader(context.Background(), stream.Writer, 10*1024, audioFile)

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
