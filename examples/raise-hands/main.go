package main

import (
	"flag"
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

	log.SetOutput(f)

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
	apiKey := os.Getenv("ZOOM_JWT_API_KEY")
	apiSecret := os.Getenv("ZOOM_JWT_API_SECRET")
	name := os.Getenv("ZOOM_NAME")
	hwid := os.Getenv("ZOOM_HWID")

	// create new session
	// meetingNumber, meetingPassword, username, hardware uuid (can be random but should be relatively constant or it will appear to zoom that you have many many many devices), proxy url, jwt api key, jwt api secret)
	session, err := zoom.NewZoomSession(meetingInfo.MeetingNumber, meetingInfo.MeetingPassword, name, hwid, "", apiKey, apiSecret)
	if err != nil {
		panic(err)
	}

	// the third argument is the "onmessage" function.  it will be triggered everytime the websocket client receives a message
	panic(session.MakeWebsocketConnection(func(session *zoom.ZoomSession, message zoom.Message) error {
		switch m := message.(type) {
		case *zoom.ConferenceRosterIndication:
			for _, person := range m.Update {
				if person.ID != session.JoinInfo.UserID {
					// Check if not nil -> update to true or false
					if person.BRaiseHand != nil {
						raisedHand := *person.BRaiseHand
						if raisedHand {
							log.Println(string(person.Dn2) + " has raised their hand!")
						} else {
							log.Println(string(person.Dn2) + " has lowered their hand!")
						}
					}
				}
			}
			return nil
		default:
			return nil
		}
	}))
}
