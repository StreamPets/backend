package announcers

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/stretchr/testify/assert"
)

func TestAddClientWithAnnouncements(t *testing.T) {
	t.Run("add client and announce join", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		pet := services.Pet{}

		announcer := NewAnnouncerService()

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.channelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Announcement{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceJoin(channelId, pet)
		wg.Wait()

		expected := Announcement{
			channelId: channelId,
			Event:     "JOIN",
			Message:   pet,
		}

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce part", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel name")
		userId := models.TwitchId("user id")

		announcer := NewAnnouncerService()

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.channelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Announcement{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnouncePart(channelId, userId)
		wg.Wait()

		expected := Announcement{
			channelId: channelId,
			Event:     "PART",
			Message:   userId,
		}

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce action", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")
		action := "action"

		announcer := NewAnnouncerService()

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.channelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Announcement{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceAction(channelId, userId, action)
		wg.Wait()

		expected := Announcement{
			channelId: channelId,
			Event:     fmt.Sprintf("%s-%s", action, userId),
			Message:   userId,
		}

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce update", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")
		image := "image"

		announcer := NewAnnouncerService()

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.channelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Announcement{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceUpdate(channelId, userId, image)
		wg.Wait()

		expected := fmt.Sprintf("%s-%s", "COLOR", userId)

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0].Event)
	})
}

func TestRemoveClientWithAnnouncements(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	pet := services.Pet{}

	announcer := NewAnnouncerService()

	client := announcer.AddClient(channelId)
	assert.Equal(t, channelId, client.channelId)

	announcer.RemoveClient(client)

	var wg sync.WaitGroup
	wg.Add(1)

	events := []Announcement{}
	go func() {
		defer wg.Done()
		for event := range client.Stream {
			events = append(events, event)
		}
	}()

	announcer.AnnounceJoin(channelId, pet)
	wg.Wait()

	if len(events) != 0 {
		t.Errorf("expected [] but got %s", events)
	}
}

func TestAnnouncerOnMultipleChannels(t *testing.T) {
	mock.SetUp(t)

	channelOneId := models.TwitchId("channel one id")
	channelTwoId := models.TwitchId("channel two id")
	pet := services.Pet{}

	announcer := NewAnnouncerService()

	clientOne := announcer.AddClient(channelOneId)
	assert.Equal(t, channelOneId, clientOne.channelId)

	clientTwo := announcer.AddClient(channelTwoId)
	assert.Equal(t, channelTwoId, clientTwo.channelId)

	var wg sync.WaitGroup
	wg.Add(1)

	eventsOne := []Announcement{}
	go func() {
		for event := range clientOne.Stream {
			eventsOne = append(eventsOne, event)
			wg.Done()
		}
	}()

	eventsTwo := []Announcement{}
	go func() {
		for event := range clientTwo.Stream {
			eventsTwo = append(eventsTwo, event)
			wg.Done()
		}
	}()

	announcer.AnnounceJoin(channelOneId, pet)
	wg.Wait()

	expected := Announcement{
		channelId: channelOneId,
		Event:     "JOIN",
		Message:   pet,
	}

	assert.Equal(t, 1, len(eventsOne))
	assert.Equal(t, expected, eventsOne[0])
	assert.Equal(t, 0, len(eventsTwo))
}
