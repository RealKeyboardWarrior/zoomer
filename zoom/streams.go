package zoom

import (
	"encoding/hex"
	"errors"
	"log"
	"time"

	"net/http"
	"net/url"

	"github.com/RealKeyboardWarrior/zoomer/zoom/codecs/opus"
	"github.com/RealKeyboardWarrior/zoomer/zoom/formats/mp4"
	"github.com/RealKeyboardWarrior/zoomer/zoom/rtp"
	"github.com/RealKeyboardWarrior/zoomer/zoom/streampkt"
	"github.com/gorilla/websocket"
)

const (
	PING                = 0x00
	RTP_SCREENSHARE_PKT = 0x4D
	// RTP_VIDEO_PKT          = 0x02 // DEPRECRATED!
	RTP_AUDIO_PKT    = 0x6B
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

func CreateZoomAudioStreams(session *ZoomSession) (*ZoomStreams, error) {
	if session.JoinInfo == nil {
		return nil, errors.New("Zoom session does not have valid JoinInfo")
	}

	if session.RwgInfo == nil {
		return nil, errors.New("Zoom session does not have valid RwgInfo")
	}

	if session.JoinInfo.ZoomID == "" {
		return nil, errors.New("Zoom session does not have valid ZoomID")
	}

	// Normal video camera sharing
	downstream := createWebSocketUrl(session, "a", "5")
	recv, err := createWebsocket("recv", downstream)
	if err != nil {
		return nil, err
	}

	// Upstream for both screenshare and audio is 2.
	// TODO: should be for video as well - verify
	upstream := createWebSocketUrl(session, "a", "2")
	send, err := createWebsocket("send", upstream)
	if err != nil {
		return nil, err
	}

	final := &ZoomStreams{
		recv:    recv,
		send:    send,
		decoder: rtp.NewZoomRtpDecoder(rtp.STREAM_TYPE_AUDIO),
	}

	go final.StartReceiveChannel()

	return final, nil
}

func CreateZoomVideoStreams(session *ZoomSession) (*ZoomStreams, error) {
	if session.JoinInfo == nil {
		return nil, errors.New("Zoom session does not have valid JoinInfo")
	}

	if session.RwgInfo == nil {
		return nil, errors.New("Zoom session does not have valid RwgInfo")
	}

	if session.JoinInfo.ZoomID == "" {
		return nil, errors.New("Zoom session does not have valid ZoomID")
	}

	// Normal video camera sharing
	downstream := createWebSocketUrl(session, "v", "5")
	recv, err := createWebsocket("recv", downstream)
	if err != nil {
		return nil, err
	}

	// Upstream for both screenshare and audio is 2.
	// TODO: should be for video as well - verify
	upstream := createWebSocketUrl(session, "v", "2")
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

func CreateZoomScreenShareStreams(session *ZoomSession) (*ZoomStreams, error) {
	if session.JoinInfo == nil {
		return nil, errors.New("Zoom session does not have valid JoinInfo")
	}

	if session.RwgInfo == nil {
		return nil, errors.New("Zoom session does not have valid RwgInfo")
	}

	if session.JoinInfo.ZoomID == "" {
		return nil, errors.New("Zoom session does not have valid ZoomID")
	}

	downstream := createWebSocketUrl(session, "s", "1")
	recv, err := createWebsocket("recv", downstream)
	if err != nil {
		return nil, err
	}

	upstream := createWebSocketUrl(session, "s", "2")
	send, err := createWebsocket("send", upstream)
	if err != nil {
		return nil, err
	}

	final := &ZoomStreams{
		recv:    recv,
		send:    send,
		decoder: rtp.NewZoomRtpDecoder(rtp.STREAM_TYPE_SCREENSHARE),
	}

	go final.StartReceiveChannel()

	return final, nil
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

	videoRecorder, err := mp4.NewRecorder()
	if err != nil {
		log.Fatal(err)
	}
	go (func() {
		time.Sleep(60 * time.Second)
		err := videoRecorder.Close()
		if err != nil {
			panic(err)
		}
	})()
	//defer videoRecorder.Close()

	audioRecorder, err := opus.CreateNewPCMRecorder()
	if err != nil {
		log.Fatal(err)
	}
	go (func() {
		time.Sleep(20 * time.Second)
		audioRecorder.Close()
	})()
	//defer audioRecorder.Close()

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
		} else if p[0] == RTP_AUDIO_PKT {
			zoomPkt := &streampkt.ZoomAudioPkt{}
			err := zoomPkt.Unmarshal(p)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Printf("%v", zoomPkt)
			sample, err := decoder.Decode(zoomPkt.Rtp)
			if err != nil {
				log.Fatal(err)
				return
			}

			if sample != nil {
				err = audioRecorder.Record(sample)
				if err != nil {
					log.Fatal(err)
					return
				}
			}
		} else if p[0] == RTP_SCREENSHARE_PKT || p[0] == RTP_VIDEO_PKT {
			log.Printf("pkt = %v", hex.EncodeToString(p))
			start := 4
			if p[0] == RTP_VIDEO_PKT {
				start = 28
			}
			sample, err := decoder.Decode(p[start:])
			if err != nil {
				log.Fatal(err)
				return
			}
			if sample != nil {
				err = videoRecorder.WritePacket(sample.Data, sample.Duration)
				if err != nil {
					log.Fatal(err)
					return
				}
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
			log.Printf("name=receive type=%v len=%v payload=%v", messageType, len(p), hex.EncodeToString(p))
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
