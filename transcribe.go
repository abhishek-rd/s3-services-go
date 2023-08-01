package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice"
)

func main() {
	var audioFiles [2]string
	audioFiles[0] = "download (2).ogg"
	audioFiles[1] = "download (1).ogg"
	for i := 0; i < len(audioFiles); i++ {
	// Create a new AWS session
	sess := session.Must(session.NewSession())

	// Create a client for the Transcribe Streaming service in the desired region
	client := transcribestreamingservice.New(sess, aws.NewConfig().WithRegion("us-west-2"))
	lc := "en-US"
	me := "ogg-opus"
	// Start the stream transcription
	resp, err := client.StartStreamTranscription(&transcribestreamingservice.StartStreamTranscriptionInput{
		LanguageCode:         &lc, //aws.String(transcribestreamingservice.LanguageCodeEnUs),
		MediaEncoding:        &me, //aws.String(transcribestreamingservice.MediaEncodingOggOpus),
		MediaSampleRateHertz: aws.Int64(16000),
	})
	if err != nil {
		log.Fatalf("failed to start streaming, %v", err)
	}

	// Get the transcription stream
	stream := resp.GetStream()

	// Close the stream when the function exits
	defer stream.Close()

	// Open the audio file
	audioFile, err := os.Open(audioFiles[i])
	if err != nil {
		log.Fatalf("failed to open audio file, %v", err)
	}
	defer audioFile.Close()

	// Stream audio from the file to the stream writer
	transcribestreamingservice.StreamAudioFromReader(context.Background(), stream.Writer, 10*1024, audioFile)
	var textout string
	// Process events from the stream
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
	log.Printf("Transcribed text: %s", textout)
}
}
