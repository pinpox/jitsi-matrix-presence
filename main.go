package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/matrix-org/gomatrix"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read and parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var hookData ProsodyHookData
	if err := json.Unmarshal(body, &hookData); err != nil {
		log.Print("JSON Unmarshal error:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	message := ""

	switch hookData.EventName {
	case "muc-occupant-joined":
		message = "someone joined the room"
	case "muc-occupant-left":
		message = "someone left the room"
	case "muc-room-created":
		message = "Room was created"
	case "muc-room-destroyed":
		message = "Room was destroyed"
	default:
		log.Printf("Unknown event-type received")
	}

	if err = sendMatrixMessage(message); err != nil {
		log.Printf("Error sending message: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send a response back to the webhook sender
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "Webhook received successfully!")
}

func getCheck(key string) string {

	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment not set: %v", key)
	}
	return val

}

var homeserverURL string
var userID string
var accessToken string
var roomID string
var listenAddress string
var matrixClient *gomatrix.Client

// var matrixClient Client

func init() {
	var err error

	homeserverURL = getCheck("HOMESERVER_URL")
	userID = getCheck("USER_ID")
	accessToken = getCheck("ACCESS_TOKEN")
	roomID = getCheck("ROOM_ID")
	listenAddress = getCheck("LISTEN_ADDRESS")

	matrixClient, err = gomatrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		panic(fmt.Errorf("failed to create matrix client: %w", err))
	}

}

func main() {
	http.HandleFunc("/webhook", webhookHandler)
	log.Printf("Starting server on %s...\n", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}

// sendMatrixMessage sends a plain-text message to the specified Matrix room.
func sendMatrixMessage(message string) error {
	var err error

	// Send a text message to the room
	_, err = matrixClient.SendText(roomID, message)
	if err != nil {
		return err
	}

	return nil
}
