package announcers

import (
	"fmt"

	"github.com/streampets/backend/pets"
)

type Announcement struct {
	Event     string
	Message   interface{}
	channelId string
}

type Client struct {
	Stream    chan Announcement
	channelId string
}

func newClient(channelId string) Client {
	return Client{channelId: channelId, Stream: make(chan Announcement)}
}

type petMap = map[string]pets.Pet
type cacheMap = map[string]petMap

func newAnnouncement(
	channelId string,
	event string,
	message interface{},
) Announcement {
	return Announcement{
		channelId: channelId,
		Event:     event,
		Message:   message,
	}
}

func joinAnnouncement(channelId string, pet pets.Pet) Announcement {
	return newAnnouncement(channelId, "JOIN", pet)
}

func partAnnouncement(channelId, userId string) Announcement {
	return newAnnouncement(channelId, "PART", userId)
}

func actionAnnouncement(channelId, userId string, action string) Announcement {
	event := fmt.Sprintf("%s-%s", action, userId)
	return newAnnouncement(channelId, event, userId)
}

func updateAnnouncement(channelId, userId string, image string) Announcement {
	event := fmt.Sprintf("COLOR-%s", userId)
	return newAnnouncement(channelId, event, image)
}
