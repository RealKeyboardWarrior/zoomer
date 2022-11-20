package zoom

import (
	"fmt"
	"testing"
)

func TestParseZoomMeetingUrl(t *testing.T) {
	meetingInfo, err := ParseZoomMeetingUrl("https://us05web.zoom.us/j/86502073975?pwd=K1V0K00wNDM1dEovVll6a2ZZcDN0UT09")
	if err != nil {
		t.Error(err)
		return
	}

	if meetingInfo.MeetingNumber != "86502073975" {
		t.Error(fmt.Errorf("unexpected meeting number"))
		return
	}
	if meetingInfo.MeetingPassword != "K1V0K00wNDM1dEovVll6a2ZZcDN0UT09" {
		t.Error(fmt.Errorf("unexpected meeting password"))
		return
	}
}
