package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/matrix-org/gomatrix"
)

var mu sync.Mutex

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

	// Parse hook data
	var hookData ProsodyHookData
	if err := json.Unmarshal(body, &hookData); err != nil {
		log.Print("JSON Unmarshal error:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Skip if room is not tracked
	if !(slices.Contains(jitsiRooms, hookData.RoomName)) {
		log.Println("Received data for untracked room:", hookData.RoomName)
		w.WriteHeader(http.StatusOK)
		return
	}

	// We need to block until the message is send, so we can save the message
	// ID. Without mutex it might send to messages instead of replacing
	mu.Lock()
	defer mu.Unlock()

	if rooms[hookData.RoomName] == nil {
		rooms[hookData.RoomName] = &RoomState{}
	}

	switch hookData.EventName {
	case "muc-occupant-joined":
		rooms[hookData.RoomName].NumParticipants = hookData.ActiveOccupantsCount
		log.Println("Got participants:", hookData.ActiveOccupantsCount)
	case "muc-occupant-left":
		if rooms[hookData.RoomName].NumParticipants == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Println("Got participants:", hookData.ActiveOccupantsCount)
		rooms[hookData.RoomName].NumParticipants = hookData.ActiveOccupantsCount
	case "muc-room-created":
		rooms[hookData.RoomName].NumParticipants = 0
	case "muc-room-destroyed":
		err = deleteMatrixMessage(rooms[hookData.RoomName].MsgID)
		delete(rooms, hookData.RoomName)
		w.WriteHeader(http.StatusOK)
		return
	default:
		log.Println("Unknown event-type received:", hookData.EventName)
	}

	message := fmt.Sprintf(
		"☎️  Call at <a href='%v/%v'>%v</a> started<br>Currently %v participant(s) in the call",
		jitsiServer, hookData.RoomName, hookData.RoomName, rooms[hookData.RoomName].NumParticipants)

	if err = sendOrUpdate(message, hookData.RoomName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to send or update message", err)
	}

	// Send a response back to the webhook sender
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "Webhook received successfully!")
}

func sendOrUpdate(message, roomName string) error {

	var err error
	var newMsgID string

	if rooms[roomName].MsgID == "" {
		log.Println("Sending new message to", roomName)
		newMsgID, err = sendMatrixMessage(message)
		if err != nil {
			return err
		}
		rooms[roomName].MsgID = newMsgID
	} else {
		log.Println("Updating message to ", roomName)
		return replaceMatrixMessage(rooms[roomName].MsgID, message)
	}

	return err
}

func getCheck(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s not set\n", key)
	}
	return val
}

// env vars
var homeserverURL string
var userID string
var accessToken string
var roomID string
var listenAddress string
var jitsiServer string
var jitsiRooms []string

var matrixClient *gomatrix.Client

type RoomState struct {
	NumParticipants int
	MsgID           string
}

var rooms = make(map[string]*RoomState)

func init() {
	var err error

	homeserverURL = getCheck("HOMESERVER_URL")
	jitsiServer = getCheck("JITSI_SERVER")
	userID = getCheck("USER_ID")
	accessToken = getCheck("ACCESS_TOKEN")
	roomID = getCheck("ROOM_ID")
	listenAddress = getCheck("LISTEN_ADDRESS")
	jitsiRooms = strings.Split(getCheck("JITSI_ROOMS"), ",")

	matrixClient, err = gomatrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		panic(fmt.Errorf("failed to create matrix client: %w", err))
	}
}

func main() {
	http.HandleFunc("/", webhookHandler)
	log.Printf("Starting server on %s...\n", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != nil {
		log.Fatalf("ListenAndServe failed: %v", err)
	}
}

// sendMatrixMessage sends a plain-text message to the specified Matrix room.
func sendMatrixMessage(message string) (string, error) {

	// Send a text message to the room
	resp, err := matrixClient.SendFormattedText(roomID, message, message)
	if err != nil {
		log.Println("Failed to send message:", err)
		return "", err
	}

	return resp.EventID, nil
}

// deleteMatrixMessage deletes (redacts) a message
func deleteMatrixMessage(eventID string) error {

	_, err := matrixClient.RedactEvent(roomID, eventID, &gomatrix.ReqRedact{})

	if err != nil {
		return fmt.Errorf("failed to redact message: %w", err)
	}

	return nil
}

// replaceMatrixMessage edits (replaces) a message
func replaceMatrixMessage(eventID, newMessage string) error {

	content := map[string]interface{}{
		"msgtype": "m.text",
		// The fallback body (what older clients see) often starts with "* " to indicate an edit:
		"body": "* " + newMessage,
		"m.new_content": map[string]interface{}{
			"body":           newMessage,
			"format":         "org.matrix.custom.html",
			"formatted_body": newMessage,
			"msgtype":        "m.text",
		},

		"m.relates_to": map[string]interface{}{
			"rel_type": "m.replace",
			"event_id": eventID,
		},
	}

	// Send the "m.room.message" event with the special content that references the original event
	_, err := matrixClient.SendMessageEvent(roomID, "m.room.message", content)
	if err != nil {
		return fmt.Errorf("failed to send replacement message: %w", err)
	}
	return nil

}
