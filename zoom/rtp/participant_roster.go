package rtp

import (
	"errors"
)

var (
	ErrParticipantExists       = errors.New("participant already exists")
	ErrParticipantMissing      = errors.New("participant is missing")
	ErrSsrcMissing             = errors.New("ssrc is missing")
	ErrSharedMeetingKeyMissing = errors.New("sharedMeetingKey is missing")
)

type Participant struct {
	userId      int
	ssrcs       []int
	secretNonce []byte
}

type ZoomParticipantRoster struct {
	sharedMeetingKey []byte
	participants     map[ /*userId*/ int]*Participant
}

func NewParticipantRoster() *ZoomParticipantRoster {
	return &ZoomParticipantRoster{
		participants: make(map[int]*Participant),
	}
}
func (roster *ZoomParticipantRoster) AddParticipant(userId int, secretNonce []byte) error {
	participant := roster.participants[userId]
	if participant != nil {
		return ErrParticipantExists
	}

	newParticipant := &Participant{
		userId:      userId,
		ssrcs:       make([]int, 0),
		secretNonce: secretNonce,
	}
	roster.participants[userId] = newParticipant

	return nil
}

func (roster *ZoomParticipantRoster) AddSsrcForParticipant(userId int, ssrc int) error {
	participant := roster.participants[userId]
	if participant == nil {
		return ErrParticipantMissing
	}

	participant.ssrcs = append(participant.ssrcs, ssrc)
	return nil
}

func (roster *ZoomParticipantRoster) GetSecretNonceForSSRC(ssrcNeedle int) ([]byte, error) {
	var secretNonce []byte
	for _, participant := range roster.participants {
		found := false
		for _, ssrcHay := range participant.ssrcs {
			if ssrcNeedle == ssrcHay {
				found = true
				secretNonce = participant.secretNonce
				break
			}
		}
		// There seems to be a relation between userId (17778240) and ssrc id (17778242)
		// so let's try converting ssrc to node id as a backup as well.
		if (participant.userId/10)*10 == (ssrcNeedle/10)*10 {
			found = true
			secretNonce = participant.secretNonce
		}

		if found {
			break
		}
	}

	if len(secretNonce) == 0 {
		return nil, ErrSsrcMissing
	}

	return secretNonce, nil
}

func (roster *ZoomParticipantRoster) SetSharedMeetingKey(sharedMeetingKey []byte) {
	roster.sharedMeetingKey = sharedMeetingKey
}

func (roster *ZoomParticipantRoster) GetSharedMeetingKey() ([]byte, error) {
	if len(roster.sharedMeetingKey) == 0 {
		return nil, ErrSharedMeetingKeyMissing
	}
	return roster.sharedMeetingKey, nil
}
