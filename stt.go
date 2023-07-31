package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "nuance.com/asr/v1"
)

const (
    // Nuance server URL.
    serverURL = "localhost:50051"

    // AUTH URL.
    authURL = "https://auth.crt.nuance.com/oauth2/token"
)

func main() {
    // Load the username and password from the JSON file.
    file, err := os.Open("credentials.json")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    var credentials map[string]string
    if err := json.NewDecoder(file).Decode(&credentials); err != nil {
        log.Fatal(err)
    }

    // Create a new gRPC client.
    conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create a new ASR client.
    client := asr.NewRecognizerClient(conn)

    // Create a new context.
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()

    // Get the access token.
    token, err := getAccessToken(ctx, client, credentials["username"], credentials["password"], authURL)
    if err != nil {
        log.Fatal(err)
    }

    // Open the audio file.
    file, err = os.Open("audio.ogg")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    // Send the audio file in chunks.
    transcribe(ctx, client, token, file)
}

func getAccessToken(ctx context.Context, client asr.RecognizerClient, username, password, authURL string) (string, error) {
    // Create a new authentication request.
    request := &asr.AuthenticateRequest{
        Username: username,
        Password: password,
    }

    // Send the authentication request.
    response, err := client.Authenticate(ctx, request)
    if err != nil {
        return "", err
    }

    // Return the access token.
    return response.AccessToken, nil
}

func transcribe(ctx context.Context, client asr.RecognizerClient, token string, file io.Reader) {
    // Create a new transcription request.
    request := &asr.StreamInputRequest{
        AudioFormat: "ogg",
        Token: token,
    }

    // Create a new writer.
    writer := client.StreamInput(ctx)

    // Send the audio file in chunks.
    for {
        chunk := make([]byte, 1024)
        n, err := file.Read(chunk)
        if err != nil {
            if err == io.EOF {
                break
            } else {
                log.Fatal(err)
            }
        }

        // Write the chunk to the writer.
        writer.Write(chunk[:n])
    }

    // Close the writer.
    writer.Close()

    // Wait for the transcription to complete.
    result, err := writer.CloseAndWait()
    if err != nil {
        log.Fatal(err)
    }

    // Print the transcription.
    fmt.Println(result.Text)
}
