package main

import (
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/RealKeyboardWarrior/zoomer/zoom"
	"github.com/joho/godotenv"
)

func main() {
	f, err := os.OpenFile(time.Now().Format("2006-01-02-15-04-05")+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	meetingUrl := flag.String("url", "", "Meeting URL")
	flag.Parse()

	meetingInfo, err := zoom.ParseZoomMeetingUrl(*meetingUrl)
	if err != nil {
		log.Fatalf("Error parsing url = %v", *meetingUrl)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// get keys from environment
	apiType := zoom.ZOOM_SDK_API_TYPE
	if os.Getenv("ZOOM_API_KEY") == "jwt" {
		apiType = zoom.ZOOM_JWT_API_TYPE
	}
	apiKey := os.Getenv("ZOOM_API_KEY")
	apiSecret := os.Getenv("ZOOM_API_SECRET")
	name := os.Getenv("ZOOM_NAME")
	hwid := os.Getenv("ZOOM_HWID")

	// create new session
	// meetingNumber, meetingPassword, username, hardware uuid (can be random but should be relatively constant or it will appear to zoom that you have many many many devices), proxy url, jwt api key, jwt api secret)
	session, err := zoom.NewZoomSession(meetingInfo.MeetingNumber, meetingInfo.MeetingPassword, name, hwid, "", apiType, apiKey, apiSecret)
	if err != nil {
		panic(err)
	}

	var streams *zoom.ZoomStreams

	// the third argument is the "onmessage" function.  it will be triggered everytime the websocket client receives a message
	panic(session.MakeWebsocketConnection(func(session *zoom.ZoomSession, message zoom.Message) error {
		switch m := message.(type) {
		case *zoom.ConferenceRosterIndication:
			// if we get an indication that someone joined the meeting, welcome them
			for _, person := range m.Add {
				if streams != nil {
					streams.AddParticipant(person.ID, person.ZoomID)
				}
			}
			for _, person := range m.Update {
				if person.AudioConnectionStatus == 2 {
					err = session.MuteUser(person.ID, true)
				}
			}
			return nil
		case *zoom.JoinConferenceResponse:
			// TODO(hackish): move this elsewhere
			streams, err = zoom.CreateZoomAudioStreams(session)
			if err != nil {
				return err
			}
			log.Printf("%v", streams)

			return nil
		case *zoom.SharingEncryptKeyIndication:
			// A1. Get sharing encryption key
			err := streams.SetSharedMeetingKey(m.EncryptKey)
			if err != nil {
				return err
			}

			// A2. Join the audio channel
			err = session.SignalAudioStatus(0, 1)
			if err != nil {
				return err
			}

			err = session.JoinAudioVoipChannel(true)
			if err != nil {
				return err
			}

			err = session.SignalAudioStatus(1, 2)
			if err != nil {
				return err
			}
			return nil
		default:
			return nil
		}
	}))
}
