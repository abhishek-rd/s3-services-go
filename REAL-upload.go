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
	"log"
	"mime/multipart"
	"os"
)

const bucket = "Bucket Name" // replace with your bucket name

type TestCase struct {
	File             string `json:"file"`
	GoldenTranscript string `json:"goldenTranscript"`
	STTResult        string `json:"sttResult,omitempty"`
}

func main() {
	r := gin.Default()
	r.POST("/upload", uploadFiles)
	r.Run(":8080") // listen and serve on :8080
}

func uploadFiles(c *gin.Context) {
	// Get the CSV file from the POST body
	csvFile, _ := c.FormFile("csvfile")
	openedCsvFile, _ := csvFile.Open()
	defer openedCsvFile.Close()

	// Get the parsed CSV file content
	parsedCsv := parseCsv(openedCsvFile)

	// Get the OGG files from the POST body
	oggFiles := c.Request.MultipartForm.File["audiofiles"]

	// Create a new AWS session
	s := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	testCases := uploadAndTranscribe(c, s, oggFiles, parsedCsv)

	// Save JSON to a local file
	saveJson(testCases)
}

func parseCsv(file multipart.File) map[string]string {
	r := csv.NewReader(file)
	parsed := make(map[string]string)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		parsed[record[0]] = record[1]
	}
	return parsed
}

func uploadAndTranscribe(c *gin.Context, s *session.Session, oggFiles []*multipart.FileHeader, parsedCsv map[string]string) []TestCase {
	svc := s3manager.NewUploader(s)
	var testCases []TestCase

	for _, file := range oggFiles {
		openedFile, _ := file.Open()
		defer openedFile.Close()

		// Transcribe before upload
		transcript := transcribeAudio(openedFile)

		// Then upload
		_, err := svc.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(file.Filename),
			Body:   openedFile,
		})

		if err != nil {
			log.Printf("upload error: %s", err)
			continue
		}

		// Add transcript to test case if filename matches
		if goldenTranscript, ok := parsedCsv[file.Filename]; ok {
			testCases = append(testCases, TestCase{
				File:             file.Filename,
				GoldenTranscript: goldenTranscript,
				STTResult:        transcript,
			})
		}
	}

	return testCases
}

func transcribeAudio(file multipart.File) string {
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

func saveJson(testCases []TestCase) {
	file, _ := os.Create("test_cases.json")
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.Encode(testCases)
}
