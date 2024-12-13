package services_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

func TestAddClientWithAnnouncements(t *testing.T) {
	t.Run("add client and announce join", func(t *testing.T) {
		channelName := "channel name"
		viewer := services.Viewer{}

		announcer := services.NewAnnouncerService()

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []services.Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceJoin(channelName, viewer)
		wg.Wait()

		expected := services.Event{Event: "JOIN", Message: viewer}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}
	})

	t.Run("add client and announce part", func(t *testing.T) {
		channelName := "channel name"
		userID := models.TwitchID("user id")

		announcer := services.NewAnnouncerService()

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []services.Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnouncePart(channelName, userID)
		wg.Wait()

		expected := services.Event{Event: "PART", Message: userID}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}
	})

	t.Run("add client and announce action", func(t *testing.T) {
		channelName := "channel name"
		userID := models.TwitchID("user id")
		action := "action"

		announcer := services.NewAnnouncerService()

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []services.Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceAction(channelName, action, userID)
		wg.Wait()

		expected := services.Event{
			Event:   fmt.Sprintf("%s-%s", action, userID),
			Message: userID,
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
		userID := models.TwitchID("user id")
		image := "image"

		announcer := services.NewAnnouncerService()

		client := announcer.AddClient(channelName)
		if client.ChannelName != channelName {
			t.Errorf("expected %s got %s", channelName, client.ChannelName)
		}

		var wg sync.WaitGroup
		wg.Add(1)

		events := []services.Event{}
		go func() {
			for event := range client.Stream {
				events = append(events, event)
				wg.Done()
			}
		}()

		announcer.AnnounceUpdate(channelName, image, userID)
		wg.Wait()

		expected := fmt.Sprintf("%s-%s", "COLOR", userID)

		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0].Event != expected {
			t.Errorf("expected %s got %s", expected, events[0].Event)
		}
	})
}

func TestRemoveClientWithAnnouncements(t *testing.T) {
	channelName := "channel name"
	viewer := services.Viewer{}

	announcer := services.NewAnnouncerService()

	client := announcer.AddClient(channelName)
	if client.ChannelName != channelName {
		t.Errorf("expected %s got %s", channelName, client.ChannelName)
	}

	announcer.RemoveClient(client)

	var wg sync.WaitGroup
	wg.Add(1)

	events := []services.Event{}
	go func() {
		defer wg.Done()
		for event := range client.Stream {
			events = append(events, event)
		}
	}()

	announcer.AnnounceJoin(channelName, viewer)
	wg.Wait()

	if len(events) != 0 {
		t.Errorf("expected [] but got %s", events)
	}
}

func TestAnnouncerOnMultipleChannels(t *testing.T) {
	channelOneName := "channel one name"
	channelTwoName := "channel two name"
	viewer := services.Viewer{}

	announcer := services.NewAnnouncerService()

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

	eventsOne := []services.Event{}
	go func() {
		for event := range clientOne.Stream {
			eventsOne = append(eventsOne, event)
			wg.Done()
		}
	}()

	eventsTwo := []services.Event{}
	go func() {
		for event := range clientTwo.Stream {
			eventsTwo = append(eventsTwo, event)
			wg.Done()
		}
	}()

	announcer.AnnounceJoin(channelOneName, viewer)
	wg.Wait()

	expected := services.Event{Event: "JOIN", Message: viewer}
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
