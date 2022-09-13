package recording

import (
	"context"
	"encoding/json"
	"github.com/keuin/slbr/danmaku"
	"github.com/keuin/slbr/danmaku/dmmsg"
	"github.com/keuin/slbr/danmaku/dmpkg"
	"github.com/keuin/slbr/logging"
	"time"
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

const (
	kHeartBeatInterval = 30 * time.Second
)

// watch monitors live room status by subscribing messages from Bilibili danmaku server,
// which talks to the client via a WebSocket or TCP connection.
// In our implementation, we use WebSocket over SSL/TLS.
// This function will return after the live is started,
// since one connection cannot receive more than one live start event.
// Error types:
// - UnrecoverableError
// - RecoverableError
// - context.Cancelled
func watch(
	ctx context.Context,
	t TaskConfig,
	url string,
	authKey string,
	liveStatusChecker func() (bool, error),
	logger logging.Logger,
) error {
	var err error

	dm := danmaku.NewDanmakuClient()

	// connect to danmaku server for live online/offline notifications
	err = dm.Connect(ctx, url)
	if err != nil {
		return NewRecoverableTaskError("failed to connect to danmaku server", err)
	}
	defer func() {
		// this operation may be time-consuming, so run in another goroutine
		go func() {
			_ = dm.Disconnect()
		}()
	}()

	// the danmaku server requires an auth token and room id when connected
	logger.Info("ws connected. Authenticating...")
	err = dm.Authenticate(t.RoomId, authKey)
	if err != nil {
		return NewUnrecoverableTaskError("authentication failed, invalid protocol", err)
	}

	// the danmaku server requires heartbeat messages every 30 seconds
	heartbeat := func() error {
		err := dm.Heartbeat()
		return err
	}

	// send initial heartbeat immediately
	err = heartbeat()
	if err != nil {
		return NewRecoverableTaskError("heartbeat failed", err)
	}

	// create heartbeat timer
	heartBeatTimer := time.NewTicker(kHeartBeatInterval)
	defer func() { heartBeatTimer.Stop() }()

	logger.Info("Checking initial live status...")
	isLiving, err := AutoRetryWithConfig[bool](ctx, logger, &t, liveStatusChecker)
	if err != nil {
		return NewRecoverableTaskError("check initial live status failed", err)
	}
	if isLiving {
		logger.Info("The live is already started. Start recording immediately.")
		return nil
	} else {
		logger.Info("The live is not started yet. Waiting...")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-heartBeatTimer.C:
			err = heartbeat()
			if err != nil {
				return NewRecoverableTaskError("heartbeat failed", err)
			}
		default:
			var msg dmpkg.DanmakuExchange
			msg, err = dm.ReadExchange()
			if err != nil {
				return NewRecoverableTaskError("failed to read exchange from server", err)
			}
			// the exchange may be compressed
			msg, err = msg.Inflate()
			if err != nil {
				return NewUnrecoverableTaskError("failed to decompress server message", err)
			}

			switch msg.Operation {
			case dmpkg.OpLayer7Data:
				//logger.Printf("server message: op %v, body %v", msg.Operation, string(msg.Body))
				var info liveInfo
				err := json.Unmarshal(msg.Body, &info)
				if err != nil {
					logger.Error("Invalid JSON: \"%v\", exchange: %v", string(msg.Body), msg)
					return NewUnrecoverableTaskError("invalid JSON response from server", err)
				}
				switch info.Command {
				case CommandLiveStart:
					return nil
				case CommandStreamPreparing:
					break
				default:
					switch info.Command {
					case "ENTRY_EFFECT":
						fallthrough
					case "ONLINE_RANK_V2":
						fallthrough
					case "ONLINE_RANK_COUNT":
						fallthrough
					case "STOP_LIVE_ROOM_LIST":
						// useless message
						fallthrough
					case "HOT_RANK_CHANGED_V2":
						// useless message
						logger.Info("Ignore message: %v", info.Command)
					case "WATCHED_CHANGE":
						// number of watched people changed
						obj, exists := info.Data["num"]
						if !exists {
							continue
						}
						viewersNum, ok := obj.(float64)
						if !ok {
							logger.Error("Cannot parse watched people number: %v", obj)
							continue
						}
						logger.Info("The number of viewers (room: %v): %v", t.RoomId, viewersNum)
					case "INTERACT_WORD":
						var raw dmmsg.RawInteractWordMessage
						err = json.Unmarshal(msg.Body, &raw)
						if err != nil {
							logger.Error("Cannot parse RawInteractWordMessage JSON: %v", err)
							continue
						}
						logger.Info("Interact word message: user: %v medal: %v",
							raw.Data.UserName, raw.Data.FansMedal.Name)
					case "DANMU_MSG":
						var raw dmmsg.RawDanMuMessage
						err = json.Unmarshal(msg.Body, &raw)
						if err != nil {
							logger.Error("Cannot parse danmaku message as JSON: %v", err)
							continue
						}
						dmm, err := dmmsg.ParseDanmakuMessage(raw)
						if err != nil {
							logger.Error("Cannot parse danmaku message JSON: %v", err)
							continue
						}
						logger.Info("Danmaku: %v", dmm.String())
					default:
						logger.Info("Ignore unhandled server message %v %v %v",
							info.Command, msg.Operation, string(msg.Body))
					}
				}
			default:
				logger.Info("Server message: %v", msg.String())
			}

		}
	}
}
