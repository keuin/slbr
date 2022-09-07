package recording

import (
	"bilibili-livestream-archiver/common"
	"bilibili-livestream-archiver/danmaku"
	"bilibili-livestream-archiver/danmaku/dmmsg"
	"bilibili-livestream-archiver/danmaku/dmpkg"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type WatcherEvent int

const (
	WatcherLiveStart WatcherEvent = 0
	WatcherLiveStop  WatcherEvent = 1
)

type liveCommand string

const (
	CommandLiveStart       = "LIVE"
	CommandStreamPreparing = "PREPARING"
)

type liveInfo struct {
	Command liveCommand            `json:"cmd"`
	Data    map[string]interface{} `json:"data"`
}

type ErrorReason int

const (
	ErrSuccess   ErrorReason = iota // no error happens, normally closed
	ErrTransport                    // I/O error, safe to retry
	ErrProtocol                     // application protocol logic error, do not retry
)

const (
	kHeartBeatInterval = 30 * time.Second
)

// watch monitors live room status by subscribing messages from Bilibili danmaku server,
// which talks to the client via a WebSocket or TCP connection.
// In our implementation, we use WebSocket over SSL/TLS.
func watch(
	ctx context.Context,
	url string,
	authKey string,
	roomId common.RoomId,
	liveStatusChecker func() (bool, error),
	chEvent chan<- WatcherEvent,
) (error, ErrorReason) {

	logger := log.Default()

	var err error

	dm := danmaku.NewDanmakuClient()
	defer func() { _ = dm.Disconnect() }()

	// connect to danmaku server for live online/offline notifications
	err = dm.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("failed to connect to danmaku server: %w", err), ErrTransport
	}
	defer func() { _ = dm.Disconnect() }()

	// the danmaku server requires an auth token and room id when connected
	logger.Println("ws connected. Authenticating...")
	err = dm.Authenticate(roomId, authKey)
	if err != nil {
		return fmt.Errorf("auth failed: %w", err), ErrProtocol
	}

	// the danmaku server requires heartbeat messages every 30 seconds
	heartbeat := func() error {
		err := dm.Heartbeat()
		return err
	}

	// send initial heartbeat immediately
	err = heartbeat()
	if err != nil {
		return err, ErrTransport
	}

	// create heartbeat timer
	heartBeatTimer := time.NewTicker(kHeartBeatInterval)
	defer func() { heartBeatTimer.Stop() }()

	logger.Println("Checking initial live status...")
	isLiving, err := liveStatusChecker()
	if err != nil {
		return fmt.Errorf("check initial live status failed: %w", err), ErrTransport
	}

	if isLiving {
		logger.Println("The live is already started. Start recording immediately.")
		chEvent <- WatcherLiveStart
	} else {
		logger.Println("The live is not started yet. Waiting...")
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ErrSuccess
		case <-heartBeatTimer.C:
			err = heartbeat()
			if err != nil {
				return fmt.Errorf("heartbeat failed: %w", err), ErrTransport
			}
		default:
			var msg dmpkg.DanmakuExchange
			msg, err = dm.ReadExchange()
			if err != nil {
				return fmt.Errorf("exchange read failed: %w", err), ErrTransport
			}
			// the exchange may be compressed
			msg, err = msg.Inflate()
			if err != nil {
				return fmt.Errorf("inflate server message failed: %v", err), ErrProtocol
			}

			switch msg.Operation {
			case dmpkg.OpLayer7Data:
				//logger.Printf("server message: op %v, body %v\n", msg.Operation, string(msg.Body))
				var info liveInfo
				err := json.Unmarshal(msg.Body, &info)
				if err != nil {
					logger.Printf("ERROR: invalid JSON: \"%v\", exchange: %v", string(msg.Body), msg)
					return fmt.Errorf("decode server message body JSON failed: %w", err), ErrProtocol
				}
				switch info.Command {
				case CommandLiveStart:
					if !isLiving {
						chEvent <- WatcherLiveStart
						isLiving = true
					}
				case CommandStreamPreparing:
					if isLiving {
						chEvent <- WatcherLiveStop
					}
				default:
					switch info.Command {
					case "ONLINE_RANK_COUNT":
						fallthrough
					case "STOP_LIVE_ROOM_LIST":
						// useless message
						fallthrough
					case "HOT_RANK_CHANGED_V2":
						// useless message
						logger.Printf("Ignore message: %v\n", info.Command)
					case "WATCHED_CHANGE":
						// number of watched people changed
						obj, exists := info.Data["num"]
						if !exists {
							continue
						}
						watchedPeopleNumber, ok := obj.(float64)
						if !ok {
							logger.Printf("Cannot parse watched people number: %v\n", obj)
							continue
						}
						logger.Printf("Watched people (room: %v): %v", roomId, watchedPeopleNumber)
					case "INTERACT_WORD":
						var raw dmmsg.RawInteractWordMessage
						err = json.Unmarshal(msg.Body, &raw)
						if err != nil {
							logger.Printf("Cannot parse RawInteractWordMessage JSON: %v\n", err)
							continue
						}
						logger.Printf("Interact word message: user: %v medal: %v",
							raw.Data.UserName, raw.Data.FansMedal.Name)
					case "DANMU_MSG":
						var raw dmmsg.RawDanMuMessage
						err = json.Unmarshal(msg.Body, &raw)
						if err != nil {
							logger.Printf("Cannot parse Dan Mu message as JSON: %v\n", err)
							continue
						}
						dmm, err := dmmsg.ParseDanmakuMessage(raw)
						if err != nil {
							logger.Printf("Cannot parse Dan Mu message JSON: %v\n", err)
							continue
						}
						logger.Printf("Dan Mu: %v\n", dmm.String())
					default:
						logger.Printf("Ignoring server message %v %v %v\n",
							info.Command, msg.Operation, string(msg.Body))
					}
				}
			default:
				logger.Printf("Server message: %v\n", msg.String())
			}

		}
	}
}
