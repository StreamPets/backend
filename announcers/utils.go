package announcers

import (
	"fmt"

	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

type Announcement struct {
	Event     string
	Message   interface{}
	channelId twitch.Id
}

type Client struct {
	Stream    chan Announcement
	channelId twitch.Id
}

func newClient(channelId twitch.Id) Client {
	return Client{channelId: channelId, Stream: make(chan Announcement)}
}

type petMap = map[twitch.Id]services.Pet
type cacheMap = map[twitch.Id]petMap

func newAnnouncement(
	channelId twitch.Id,
	event string,
	message interface{},
) Announcement {
	return Announcement{
		channelId: channelId,
		Event:     event,
		Message:   message,
	}
}

func joinAnnouncement(channelId twitch.Id, pet services.Pet) Announcement {
	return newAnnouncement(channelId, "JOIN", pet)
}

func partAnnouncement(channelId, userId twitch.Id) Announcement {
	return newAnnouncement(channelId, "PART", userId)
}

func actionAnnouncement(channelId, userId twitch.Id, action string) Announcement {
	event := fmt.Sprintf("%s-%s", action, userId)
	return newAnnouncement(channelId, event, userId)
}

func updateAnnouncement(channelId, userId twitch.Id, image string) Announcement {
	event := fmt.Sprintf("COLOR-%s", userId)
	return newAnnouncement(channelId, event, image)
}
