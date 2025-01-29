package services

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestAddClientWithAnnouncements(t *testing.T) {
	t.Run("add client and announce join", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		pet := Pet{}

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.ChannelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceJoin(channelId, pet)
		wg.Wait()

		expected := Event{Event: "JOIN", Message: pet}

		mock.Verify(cacheMock, mock.Once()).AddPet(channelId, pet)

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce part", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel name")
		userId := models.TwitchId("user id")

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.ChannelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnouncePart(channelId, userId)
		wg.Wait()

		expected := Event{Event: "PART", Message: userId}

		mock.Verify(cacheMock, mock.Once()).RemovePet(channelId, userId)

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce action", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")
		action := "action"

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.ChannelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceAction(channelId, userId, action)
		wg.Wait()

		expected := Event{
			Event:   fmt.Sprintf("%s-%s", action, userId),
			Message: userId,
		}

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0])
	})

	t.Run("add client and announce update", func(t *testing.T) {
		mock.SetUp(t)

		channelId := models.TwitchId("channel id")
		userId := models.TwitchId("user id")
		image := "image"

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelId)
		assert.Equal(t, channelId, client.ChannelId)

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceUpdate(channelId, userId, image)
		wg.Wait()

		expected := fmt.Sprintf("%s-%s", "COLOR", userId)

		mock.Verify(cacheMock, mock.Once()).UpdatePet(channelId, userId, image)

		assert.Equal(t, 1, len(events))
		assert.Equal(t, expected, events[0].Event)
	})
}

func TestRemoveClientWithAnnouncements(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	pet := Pet{}

	cacheMock := mock.Mock[PetCache]()
	announcer := NewAnnouncerService(cacheMock)

	client := announcer.AddClient(channelId)
	assert.Equal(t, channelId, client.ChannelId)

	announcer.RemoveClient(client)

	var wg sync.WaitGroup
	wg.Add(1)

	events := []Event{}
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
	pet := Pet{}

	cacheMock := mock.Mock[PetCache]()
	announcer := NewAnnouncerService(cacheMock)

	clientOne := announcer.AddClient(channelOneId)
	assert.Equal(t, channelOneId, clientOne.ChannelId)

	clientTwo := announcer.AddClient(channelTwoId)
	assert.Equal(t, channelTwoId, clientTwo.ChannelId)

	var wg sync.WaitGroup
	wg.Add(1)

	eventsOne := []Event{}
	go func() {
		for event := range clientOne.Stream {
			eventsOne = append(eventsOne, event)
			wg.Done()
		}
	}()

	eventsTwo := []Event{}
	go func() {
		for event := range clientTwo.Stream {
			eventsTwo = append(eventsTwo, event)
			wg.Done()
		}
	}()

	announcer.AnnounceJoin(channelOneId, pet)
	wg.Wait()

	expected := Event{Event: "JOIN", Message: pet}

	assert.Equal(t, 1, len(eventsOne))
	assert.Equal(t, expected, eventsOne[0])
	assert.Equal(t, 0, len(eventsTwo))
}

func TestAddClient(t *testing.T) {
	mock.SetUp(t)

	channelId := models.TwitchId("channel id")
	pets := []Pet{{}, {}}

	cacheMock := mock.Mock[PetCache]()
	mock.When(cacheMock.GetPets(channelId)).ThenReturn(pets)

	announcer := NewAnnouncerService(cacheMock)
	client := announcer.AddClient(channelId)

	got := []Pet{}
	var wg sync.WaitGroup
	wg.Add(len(pets))

	go func() {
		for event := range client.Stream {
			pet, ok := event.Message.(Pet)

			assert.True(t, ok)

			got = append(got, pet)
			wg.Done()
		}
	}()

	wg.Wait()
	close(client.Stream)

	assert.Equal(t, pets, got)
}
