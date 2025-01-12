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
| `JITSI_ROOMS`    | Comma-separated list of jitsi rooms to tack |
| `JITSI_SERVER`   | Jitsi Server URI for the links              |

## Testing

There are examples of the data send by the webhooks in the `testdata` folder.
Use the following command to test:

```sh
curl -X POST \
     -H "Content-Type: application/json" \
     -d @testdata/muc-occupant-joined.json \
     http://localhost:8080/
```
