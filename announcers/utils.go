package announcers

import (
	"fmt"

	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/twitch"
)

type Announcement struct {
	Event     string
	Message   interface{}
	channelId twitch.UserId
}

type Client struct {
	Stream    chan Announcement
	channelId twitch.UserId
}

func newClient(channelId twitch.UserId) Client {
	return Client{channelId: channelId, Stream: make(chan Announcement)}
}

type petMap = map[twitch.UserId]pets.Pet
type cacheMap = map[twitch.UserId]petMap

func newAnnouncement(
	channelId twitch.UserId,
	event string,
	message interface{},
) Announcement {
	return Announcement{
		channelId: channelId,
		Event:     event,
		Message:   message,
	}
}

func joinAnnouncement(channelId twitch.UserId, pet pets.Pet) Announcement {
	return newAnnouncement(channelId, "JOIN", pet)
}

func partAnnouncement(channelId, userId twitch.UserId) Announcement {
	return newAnnouncement(channelId, "PART", userId)
}

func actionAnnouncement(channelId, userId twitch.UserId, action string) Announcement {
	event := fmt.Sprintf("%s-%s", action, userId)
	return newAnnouncement(channelId, event, userId)
}

func updateAnnouncement(channelId, userId twitch.UserId, image string) Announcement {
	event := fmt.Sprintf("COLOR-%s", userId)
	return newAnnouncement(channelId, event, image)
}
