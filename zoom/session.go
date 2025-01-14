package zoom

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ZoomApiType string

const (
	ZOOM_JWT_API_TYPE ZoomApiType = "jwt"
	ZOOM_SDK_API_TYPE ZoomApiType = "sdk"
)

type ZoomSession struct {
	mu sync.Mutex

	MeetingNumber   string
	MeetingPassword string
	Username        string
	HardwareID      uuid.UUID
	ZoomApiType     ZoomApiType
	ZoomApiKey      string
	ZoomApiSecret   string
	JoinInfo        *JoinConferenceResponse
	ProxyURL        *url.URL

	meetingOpt          string
	httpClient          *http.Client
	websocketConnection *websocket.Conn
	sendSequenceNumber  uint32

	// RWG
	RwgInfo   *RwgInfo
	RwgCookie string
}

func NewZoomSession(meetingNumber string, meetingPassword string, username string, hardwareID string, proxyURL string, zoomApiType ZoomApiType, zoomApiKey string, zoomApiSecret string) (*ZoomSession, error) {
	if meetingNumber == "" || meetingPassword == "" || username == "" || hardwareID == "" || zoomApiKey == "" || zoomApiSecret == "" {
		return nil, errors.New("Please make sure to provide values for meeting number, meeting password, username, hardware ID (hardware ID must be in the format of UUID), and API key/secret.")
	}
	uuidParsed, err := uuid.Parse(hardwareID)
	if err != nil {
		return nil, err
	}
	session := ZoomSession{
		MeetingNumber:   strings.Replace(meetingNumber, " ", "", -1), // remove all
		MeetingPassword: meetingPassword,
		Username:        username,
		HardwareID:      uuidParsed,
		ZoomApiType:     zoomApiType,
		ZoomApiKey:      zoomApiKey,
		ZoomApiSecret:   zoomApiSecret,
	}

	session.httpClient = &http.Client{
		Timeout: 35 * time.Second, // largeish timeout for slow proxies
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // ignore certificate errors so we can use charles to debug
		},
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	if proxyURL != "" {
		proxyUrlParsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		// save for websocket client later
		session.ProxyURL = proxyUrlParsed
		transport.Proxy = http.ProxyURL(session.ProxyURL)
	}
	session.httpClient.Transport = transport

	return &session, nil
}
