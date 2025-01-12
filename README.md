# Jitsi Matrix Presence Notificator

Will send a message to a Matrix channel every time users join or leave Jitsi

## Configuration

Configuration is done via environment variables

| Variable         | Description                                 |
|------------------|---------------------------------------------|
| `HOMESERVER_URL` | Homeserver URL                              |
| `USER_ID`        | UserID to post as                           |
| `ACCESS_TOKEN`   | Typically obtained via login or auth method |
| `ROOM_ID`        | e.g. "!YourRoomID:matrix.example.com"       |
| `LISTEN_ADDRESS` | Adress to listen for Webhooks               |
