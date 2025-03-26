package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/labstack/echo/v4"
)

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
	e.POST("/calendar", sendDataMessage)
	e.POST("/notification", sendNotification)
	e.Logger.Fatal(e.Start(":3000"))
}
