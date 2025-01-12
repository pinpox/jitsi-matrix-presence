package main

type ProsodyHookData struct {
	ActiveOccupantsCount int        `json:"active_occupants_count"`
	CreatedAt            int        `json:"created_at"`
	DestroyedAt          int        `json:"destroyed_at"`
	EventName            string     `json:"event_name"`
	IsBreakout           bool       `json:"is_breakout"`
	RoomJid              string     `json:"room_jid"`
	RoomName             string     `json:"room_name"`
	AllOccupants         []Occupant `json:"all_occupants"`
	Occupant             Occupant   `json:"occupant"`
}

type Occupant struct {
	Email       string `json:"email"`
	ID          string `json:"id"`
	JoinedAt    int    `json:"joined_at"`
	LeftAt      int    `json:"left_at"`
	Name        string `json:"name"`
	OccupantJid string `json:"occupant_jid"`
}
