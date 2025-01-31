package announcers

import (
	"fmt"

	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type Announcement struct {
	Event     string
	Message   interface{}
	channelId models.TwitchId
}

type Client struct {
	Stream    chan Announcement
	channelId models.TwitchId
}

func newClient(channelId models.TwitchId) *Client {
	return &Client{channelId: channelId, Stream: make(chan Announcement)}
}

type petMap = map[models.TwitchId]services.Pet
type cacheMap = map[models.TwitchId]petMap

func newAnnouncement(
	channelId models.TwitchId,
	event string,
	message interface{},
) Announcement {
	return Announcement{
		channelId: channelId,
		Event:     event,
		Message:   message,
	}
}

func joinAnnouncement(channelId models.TwitchId, pet services.Pet) Announcement {
	return newAnnouncement(channelId, "JOIN", pet)
}

func partAnnouncement(channelId, userId models.TwitchId) Announcement {
	return newAnnouncement(channelId, "PART", userId)
}

func actionAnnouncement(channelId, userId models.TwitchId, action string) Announcement {
	event := fmt.Sprintf("%s-%s", action, userId)
	return newAnnouncement(channelId, event, userId)
}

func updateAnnouncement(channelId, userId models.TwitchId, image string) Announcement {
	event := fmt.Sprintf("COLOR-%s", userId)
	return newAnnouncement(channelId, event, image)
}
