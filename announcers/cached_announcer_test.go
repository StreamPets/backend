package announcers

import (
	"sync"
	"testing"
	"time"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestAddClient(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	expected := newClient(channelId)

	announcerMock := mock.Mock[announcer]()
	mock.When(announcerMock.AddClient(channelId)).ThenReturn(expected)

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	actual := cachedAnnouncer.AddClient(channelId)

	assert.Equal(t, expected, actual)
}

func TestRemoveClient(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	client := newClient(channelId)

	announcerMock := mock.Mock[announcer]()

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	cachedAnnouncer.RemoveClient(client)

	mock.Verify(announcerMock, mock.Once()).RemoveClient(client)
}

func TestAnnounceJoin(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")

	pet := services.Pet{}
	client := newClient(channelId)

	announcerMock := mock.Mock[announcer]()
	mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	cachedAnnouncer.AnnounceJoin(channelId, pet)
	cachedAnnouncer.AddClient(channelId)

	var wg sync.WaitGroup
	wg.Add(1)

	announcements := []Announcement{}
	go func() {
		for announcement := range client.Stream {
			announcements = append(announcements, announcement)
			wg.Done()
		}
	}()

	wg.Wait()

	expected := joinAnnouncement(channelId, pet)

	assert.Equal(t, 1, len(announcements))
	assert.Equal(t, expected, announcements[0])

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelId, pet)
}

func TestAnnouncePart(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")

	pet := services.Pet{UserId: userId}
	client := newClient(channelId)

	announcerMock := mock.Mock[announcer]()
	mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	cachedAnnouncer.AnnounceJoin(channelId, pet)
	cachedAnnouncer.AnnouncePart(channelId, userId)
	cachedAnnouncer.AddClient(channelId)

	select {
	case msg := <-client.Stream:
		t.Errorf("did not expect a msg but received %s", msg)
	case <-time.After(1 * time.Second):
	}

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelId, userId)
}

func TestAnnounceAction(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")
	action := "action"

	announcerMock := mock.Mock[announcer]()

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	cachedAnnouncer.AnnounceAction(channelId, userId, action)

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelId, userId, action)
}

func TestAnnounceUpdate(t *testing.T) {
	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")
	image := "image"
	newImage := "new image"

	pet := services.Pet{UserId: userId, Image: image}
	client := newClient(channelId)

	announcerMock := mock.Mock[announcer]()
	mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

	cachedAnnouncer := NewCachedAnnouncerService(announcerMock)
	cachedAnnouncer.AnnounceJoin(channelId, pet)
	cachedAnnouncer.AnnounceUpdate(channelId, userId, newImage)
	cachedAnnouncer.AddClient(channelId)

	var wg sync.WaitGroup
	wg.Add(1)

	announcements := []Announcement{}
	go func() {
		for announcement := range client.Stream {
			announcements = append(announcements, announcement)
			wg.Done()
		}
	}()

	wg.Wait()

	assert.Equal(t, 1, len(announcements))

	actual := announcements[0].Message.(services.Pet)
	expected := services.Pet{UserId: userId, Image: newImage}

	assert.Equal(t, expected, actual)

	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelId, userId, newImage)
}
