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

	log.Println(string(body))

	var hookData ProsodyHookData
	if err := json.Unmarshal(body, &hookData); err != nil {
		log.Print("JSON Unmarshal error:", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	message := ""

	switch hookData.EventName {
	case "muc-occupant-joined":
		rooms.Join(hookData.RoomName, hookData.Occupant.ID)
	//	// We only notify on first person joining
	//	if len(rooms[hookData.RoomName]) == 1 {
	//		if hookData.Occupant.Name == "" {
	//			hookData.Occupant.Name = "Someone"
	//		}
	//		message = fmt.Sprintf("☎️  %v joined #%v", hookData.Occupant.Name, hookData.RoomName)
	//	}
	case "muc-occupant-left":
		rooms.Leave(hookData.RoomName, hookData.Occupant.ID)
		// No notification needed. Call will end when last person leaves
		// message = fmt.Sprintf("%☎️%v left #%v", hookData.Occupant.Name, hookData.RoomName)
	case "muc-room-created":
		rooms.Create(hookData.RoomName)
		message = fmt.Sprintf("☎️  Call at <a href='%v/%v'>%v</a> started", jitsiServer, hookData.RoomName, hookData.RoomName)
	case "muc-room-destroyed":
		rooms.Destroy(hookData.RoomName)
		message = fmt.Sprintf("☎️  Call at <a href='%v/%v'>%v</a> ended", jitsiServer, hookData.RoomName, hookData.RoomName)
	default:
		log.Println("Unknown event-type received:", hookData.EventName)
	}

	if slices.Contains(jitsiRooms, hookData.RoomName) && message != "" {

		if err = sendMatrixMessage(message); err != nil {
			log.Printf("Error sending message: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Send a response back to the webhook sender
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "Webhook received successfully!")
}

type Rooms map[string][]string

var rooms Rooms = make(map[string][]string)

func (rs Rooms) Create(room string) {
	rs[room] = []string{}
}

func (rs Rooms) Destroy(room string) {
	delete(rs, room)
}

func (rs Rooms) Join(room, user string) {
	rs[room] = append(rs[room], user)
}

func (rs Rooms) Leave(room, user string) {
	r := rs[room]
	for i, u := range r {
		if u == user {
			rs[room] = append(r[:i], r[i+1:]...)
		}
	}
}

func getCheck(key string) string {

	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s not set\n", key)
	}

	return val
}

var homeserverURL string
var userID string
var accessToken string
var roomID string
var listenAddress string
var jitsiServer string
var jitsiRooms []string
var matrixClient *gomatrix.Client

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
func sendMatrixMessage(message string) error {
	var err error

	// Send a text message to the room
	_, err = matrixClient.SendFormattedText(roomID, message, message)
	if err != nil {
		return err
	}

	return nil
}
