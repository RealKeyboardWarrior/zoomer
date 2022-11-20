package zoom

import (
	"fmt"
	"net/url"
	"strings"
)

type ZoomMeetingInfo struct {
	MeetingNumber   string
	MeetingPassword string
}

func ParseZoomMeetingUrl(meetingUrl string) (*ZoomMeetingInfo, error) {
	u, err := url.Parse(meetingUrl)
	if err != nil {
		return nil, err
	}

	m, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}

	splitted := strings.Split(u.Path, "/")
	if len(splitted) != 3 {
		return nil, fmt.Errorf("expected path in meeting url to be /j/18456188")
	}

	return &ZoomMeetingInfo{
		MeetingNumber:   splitted[2],
		MeetingPassword: m["pwd"][0],
	}, nil
}
