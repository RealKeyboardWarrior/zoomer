package zoom

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"net/http"
	"net/url"

	"github.com/RealKeyboardWarrior/zoomer/zoom/rtp"
	"github.com/gorilla/websocket"
)

const (
	PING                = 0x00
	RTP_SCREENSHARE_PKT = 0x4D
	// RTP_VIDEO_PKT          = 0x02 // DEPRECRATED!
	RTP_VIDEO_PKT    = 0x67
	RTCP             = 0x4E
	AES_GCM_IV_VALUE = 0x42
)

type ZoomStreams struct {
	recv *websocket.Conn
	send *websocket.Conn

	decoder *rtp.ZoomRtpDecoder
}

func createWebSocketUrl(session *ZoomSession, subType string, mode string) string {
	values := url.Values{}
	values.Set("type", subType)
	values.Set("cid", session.JoinInfo.ConID)
	values.Set("mode", mode)
	url := &url.URL{
		Scheme:   "wss",
		Host:     session.RwgInfo.Rwg,
		Path:     "/wc/media/" + session.MeetingNumber,
		RawQuery: values.Encode(),
	}
	return url.String()
}

func CreateZoomStreams(session *ZoomSession, isScreenShare bool) (*ZoomStreams, error) {
	if session.JoinInfo == nil {
		return nil, errors.New("Zoom session does not have valid JoinInfo")
	}

	if session.RwgInfo == nil {
		return nil, errors.New("Zoom session does not have valid RwgInfo")
	}

	if session.JoinInfo.ZoomID == "" {
		return nil, errors.New("Zoom session does not have valid ZoomID")
	}

	/*
		secretNonce, err := ZoomEscapedBase64Decode(session.JoinInfo.ZoomID)
		if err != nil {
			return nil, err
		}
	*/

	var downstream string
	if isScreenShare {
		// Screen sharing
		downstream = createWebSocketUrl(session, "s", "1")
	} else {
		// Normal video camera sharing
		downstream = createWebSocketUrl(session, "v", "5")
	}
	recv, err := createWebsocket("recv", downstream)
	if err != nil {
		return nil, err
	}

	// Upstream for both screenshare and audio is 2.
	// TODO: should be for video as well - verify
	var upstream string
	if isScreenShare {
		// Screen sharing
		upstream = createWebSocketUrl(session, "s", "2")
	} else {
		// Normal video camera sharing
		upstream = createWebSocketUrl(session, "v", "2")
	}
	send, err := createWebsocket("send", upstream)
	if err != nil {
		return nil, err
	}

	final := &ZoomStreams{
		recv:    recv,
		send:    send,
		decoder: rtp.NewZoomRtpDecoder(rtp.STREAM_TYPE_VIDEO),
	}

	go final.StartReceiveChannel()

	return final, nil
}

// wss://zoomamn89rwg.am.zoom.us/wc/media/87144340981?type=s&cid=784C146F-A782-A49B-4487-0148303B7D86&mode=1

func Recorder() (io.WriteCloser, error) {
	f, err := os.Create(time.Now().Format("2006-01-02-15-04-05") + ".h264")
	if err != nil {
		return nil, err
	}
	return f, nil
	// defer f.Close()
	// n2, err := f.Write(d2)
}

func createWebsocket(name string, websocketUrl string) (*websocket.Conn, error) {
	log.Printf("CreateZoomStreams: dialing url= %v", websocketUrl)
	dialer := websocket.Dialer{
		EnableCompression: true,
	}

	websocketHeaders := http.Header{}
	websocketHeaders.Set("Accept-Language", "en-US,en;q=0.9")
	websocketHeaders.Set("Cache-Control", "no-cache")
	websocketHeaders.Set("Origin", "https://zoom.us")
	websocketHeaders.Set("Pragma", "no-cache")
	websocketHeaders.Set("User-Agent", userAgent)

	connection, _, err := dialer.Dial(websocketUrl, websocketHeaders)
	if err != nil {
		return nil, err
	}

	log.Printf("Dialed : %v", websocketUrl)

	return connection, nil
}

func (streams *ZoomStreams) StartReceiveChannel() {
	connection := streams.recv

	closeHandler := func(i int, msg string) error {
		log.Printf("Closing : %v %v", i, msg)
		return nil
	}
	connection.SetCloseHandler(closeHandler)

	recorder, err := Recorder()
	if err != nil {
		log.Fatal(err)
	}
	defer recorder.Close()

	decoder := streams.decoder

	for {
		messageType, p, err := connection.ReadMessage()
		if err != nil {
			log.Fatal(err)
			return
		}

		// Pong
		if p[0] == PING {
			err := connection.WriteMessage(websocket.BinaryMessage, p)
			if err != nil {
				log.Fatal(err)
				return
			}
			// RTP packet
		} else if p[0] == RTP_SCREENSHARE_PKT || p[0] == RTP_VIDEO_PKT {
			log.Printf("pkt = %v", hex.EncodeToString(p))
			start := 4
			if p[0] == RTP_VIDEO_PKT {
				start = 28
			}
			frame, err := decoder.Decode(p[start:])
			if err != nil {
				log.Fatal(err)
				return
			}
			_, err = recorder.Write(frame)
			if err != nil {
				log.Fatal(err)
				return
			}
		} else if p[0] == RTCP {
			_, err := rtp.RtcpProcess(p[4:])
			if err != nil {
				log.Fatal(err)
				return
			}
		} else if p[0] == AES_GCM_IV_VALUE {
			// log.Printf("AES_GCM_IV_VALUE IV=%v", p[4:])
		} else {
			log.Printf("name=receive type=%v len=%v payload=%v", messageType, len(p), p)
		}
	}

}

func (streams *ZoomStreams) SetSharedMeetingKey(encryptionKey string) error {
	sharedMeetingKey, err := ZoomEscapedBase64Decode(encryptionKey)
	if err != nil {
		return err
	}

	streams.decoder.ParticipantRoster.SetSharedMeetingKey(sharedMeetingKey)
	return nil
}

func (streams *ZoomStreams) AddParticipant(userId int, zoomId string) error {
	secretNonce, err := ZoomEscapedBase64Decode(zoomId)
	if err != nil {
		return err
	}

	streams.decoder.ParticipantRoster.AddParticipant(userId, secretNonce)
	return nil
}

func (streams *ZoomStreams) AddSsrcForParticipant(userId int, ssrc int) error {
	return streams.AddSsrcForParticipant(userId, ssrc)
}
