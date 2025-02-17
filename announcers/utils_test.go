package announcers

import (
	"testing"

	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestJoinAnnouncement(t *testing.T) {
	channelId := twitch.UserId("channel id")
	pet := pets.Pet{}

	actual := joinAnnouncement(channelId, pet)
	expected := Announcement{
		channelId: channelId,
		Event:     "JOIN",
		Message:   pet,
	}

	assert.Equal(t, expected, actual)
}

func TestPartAnnouncement(t *testing.T) {
	channelId := twitch.UserId("channel id")
	userId := twitch.UserId("user id")

	actual := partAnnouncement(channelId, userId)
	expected := Announcement{
		channelId: channelId,
		Event:     "PART",
		Message:   userId,
	}

	assert.Equal(t, expected, actual)
}

func TestActionAnnouncement(t *testing.T) {
	channelId := twitch.UserId("channel id")
	userId := twitch.UserId("user id")
	action := "action"

	actual := actionAnnouncement(channelId, userId, action)
	expected := Announcement{
		channelId: channelId,
		Event:     "action-user id",
		Message:   userId,
	}

	assert.Equal(t, expected, actual)
}

func TestUpdateAnnouncement(t *testing.T) {
	channelId := twitch.UserId("channel id")
	userId := twitch.UserId("user id")
	image := "image"

	actual := updateAnnouncement(channelId, userId, image)
	expected := Announcement{
		channelId: channelId,
		Event:     "COLOR-user id",
		Message:   image,
	}

	assert.Equal(t, expected, actual)
}
