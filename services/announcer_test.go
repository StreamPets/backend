package services

import (
	"fmt"
	"slices"
	"sync"
	"testing"

	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
)

func TestAddClientWithAnnouncements(t *testing.T) {
	t.Run("add client and announce join", func(t *testing.T) {
		channelName := "channel name"
		pet := Pet{}

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceJoin(channelName, pet)
		wg.Wait()

		expected := Event{Event: "JOIN", Message: pet}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}

		mock.Verify(cacheMock, mock.Once()).AddPet(channelName, pet)
	})

	t.Run("add client and announce part", func(t *testing.T) {
		channelName := "channel name"
		userId := models.TwitchId("user id")

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnouncePart(channelName, userId)
		wg.Wait()

		expected := Event{Event: "PART", Message: userId}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}

		mock.Verify(cacheMock, mock.Once()).RemovePet(channelName, userId)
	})

	t.Run("add client and announce action", func(t *testing.T) {
		channelName := "channel name"
		userId := models.TwitchId("user id")
		action := "action"

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceAction(channelName, action, userId)
		wg.Wait()

		expected := Event{
			Event:   fmt.Sprintf("%s-%s", action, userId),
			Message: userId,
		}

		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}
	})

	t.Run("add client and announce update", func(t *testing.T) {
		channelName := "channel name"
		userId := models.TwitchId("user id")
		image := "image"

		cacheMock := mock.Mock[PetCache]()
		announcer := NewAnnouncerService(cacheMock)

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceUpdate(channelName, image, userId)
		wg.Wait()

		expected := fmt.Sprintf("%s-%s", "COLOR", userId)

		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0].Event != expected {
			t.Errorf("expected %s got %s", expected, events[0].Event)
		}

		mock.Verify(cacheMock, mock.Once()).UpdatePet(channelName, image, userId)
	})
}

func TestRemoveClientWithAnnouncements(t *testing.T) {
	channelName := "channel name"
	pet := Pet{}

	cacheMock := mock.Mock[PetCache]()
	announcer := NewAnnouncerService(cacheMock)

	client := announcer.AddClient(channelName)
	if client.ChannelName != channelName {
		t.Errorf("expected %s got %s", channelName, client.ChannelName)
	}

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

	announcer.AnnounceJoin(channelName, pet)
	wg.Wait()

	if len(events) != 0 {
		t.Errorf("expected [] but got %s", events)
	}
}

func TestAnnouncerOnMultipleChannels(t *testing.T) {
	channelOneName := "channel one name"
	channelTwoName := "channel two name"
	pet := Pet{}

	cacheMock := mock.Mock[PetCache]()
	announcer := NewAnnouncerService(cacheMock)

	clientOne := announcer.AddClient(channelOneName)
	if clientOne.ChannelName != channelOneName {
		t.Errorf("expected %s got %s", channelOneName, clientOne.ChannelName)
	}

	clientTwo := announcer.AddClient(channelTwoName)
	if clientTwo.ChannelName != channelTwoName {
		t.Errorf("expected %s got %s", channelTwoName, clientTwo.ChannelName)
	}

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

	announcer.AnnounceJoin(channelOneName, pet)
	wg.Wait()

	expected := Event{Event: "JOIN", Message: pet}
	if len(eventsOne) != 1 {
		t.Errorf("expected 1 event but got %d", len(eventsOne))
	}
	if eventsOne[0] != expected {
		t.Errorf("expected %s got %s", expected, eventsOne[0])
	}
	if len(eventsTwo) != 0 {
		t.Errorf("expected [] but got %s", eventsTwo)
	}
}

func TestAddClient(t *testing.T) {
	channelName := "channel name"
	pets := []Pet{{}, {}}

	cacheMock := mock.Mock[PetCache]()
	mock.When(cacheMock.GetPets(channelName)).ThenReturn(pets)

	announcer := NewAnnouncerService(cacheMock)
	client := announcer.AddClient(channelName)

	got := []Pet{}
	var wg sync.WaitGroup
	wg.Add(len(pets))

	go func() {
		for event := range client.Stream {
			pet, ok := event.Message.(Pet)
			if !ok {
				t.Errorf("expected a pet but got %s", event.Message)
			}
			got = append(got, pet)
			wg.Done()
		}
	}()

	wg.Wait()
	close(client.Stream)

	if !slices.Equal(got, pets) {
		t.Errorf("got %s want %s", got, pets)
	}
}
