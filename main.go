package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2/google"
)

func generateScopedToken(ctx context.Context, scopes []string) (string, error) {
	// This uses Application Default Credentials
	credentials, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return "", fmt.Errorf("failed to get credentials: %v", err)
	}

	// Create a token source from credentials
	tokenSource := credentials.TokenSource

	// Get a token
	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	return token.AccessToken, nil
}

type WatchRequest struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Address string         `json:"address"`
	Params  map[string]int `json:"params"`
}

func watchCalendarEvents(c echo.Context) error {
	ctx := context.Background()
	token, err := generateScopedToken(
		ctx,
		[]string{
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/calendar",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate scoped token: %v", err)
	}

	watchReq := WatchRequest{
		ID:      "ccc-mobile-alerts",
		Type:    "web_hook",
		Address: "https://www.googleapis.com/auth/calendar.events",
		Params: map[string]int{
			"ttl": 3600,
		},
	}

	// Convert the request to JSON
	payload, err := json.Marshal(watchReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create the HTTP request
	url := "https://www.googleapis.com/calendar/v3/calendars/cruceschessclub@gmail.com/events/watch"
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read and check the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Successfully created calendar watch: %s\n", string(body))
	return nil
}

func sendDataMessage(c echo.Context) error {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "cruceschessclub-d09a7",
	})
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	// Obtain a messaging.Client from the App.
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
		return err
	}

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data:  map[string]string{},
		Topic: "calendar",
		Android: &messaging.AndroidConfig{
			Priority: "normal",
		},
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	return nil
}

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func sendNotification(c echo.Context) error {
	var notification Notification
	// Get the data from the request
	requestBody := c.Request().Body
	defer requestBody.Close()
	err := json.NewDecoder(requestBody).Decode(&notification)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}
	fmt.Println(notification)

	ctx := context.Background()
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "cruceschessclub-d09a7",
	})
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	// Obtain a messaging.Client from the App.
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
		return err
	}

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data: map[string]string{},
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Topic: "calendar",
		Android: &messaging.AndroidConfig{
			Priority: "normal",
		},
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	return nil
}

func main() {
	e := echo.New()
	e.GET("/", sendDataMessage)
	e.GET("/watch", watchCalendarEvents)
	e.POST("/calendar", sendDataMessage)
	e.POST("/notification", sendNotification)
	e.Logger.Fatal(e.Start(":3000"))
}
