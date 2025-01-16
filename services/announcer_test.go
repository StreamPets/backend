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
		viewer := Pet{}

		cacheMock := mock.Mock[ViewerCache]()
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

		announcer.AnnounceJoin(channelName, viewer)
		wg.Wait()

		expected := Event{Event: "JOIN", Message: viewer}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}

		mock.Verify(cacheMock, mock.Once()).AddViewer(channelName, viewer)
	})

	t.Run("add client and announce part", func(t *testing.T) {
		channelName := "channel name"
		viewerId := models.TwitchId("viewer id")

		cacheMock := mock.Mock[ViewerCache]()
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

		announcer.AnnouncePart(channelName, viewerId)
		wg.Wait()

		expected := Event{Event: "PART", Message: viewerId}
		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0] != expected {
			t.Errorf("expected %s got %s", expected, events[0])
		}

		mock.Verify(cacheMock, mock.Once()).RemoveViewer(channelName, viewerId)
	})

	t.Run("add client and announce action", func(t *testing.T) {
		channelName := "channel name"
		viewerId := models.TwitchId("viewer id")
		action := "action"

		cacheMock := mock.Mock[ViewerCache]()
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

		announcer.AnnounceAction(channelName, action, viewerId)
		wg.Wait()

		expected := Event{
			Event:   fmt.Sprintf("%s-%s", action, viewerId),
			Message: viewerId,
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
		viewerId := models.TwitchId("viewer id")
		image := "image"

		cacheMock := mock.Mock[ViewerCache]()
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

		announcer.AnnounceUpdate(channelName, image, viewerId)
		wg.Wait()

		expected := fmt.Sprintf("%s-%s", "COLOR", viewerId)

		if len(events) != 1 {
			t.Errorf("expected 1 event but got %d", len(events))
		}
		if events[0].Event != expected {
			t.Errorf("expected %s got %s", expected, events[0].Event)
		}

		mock.Verify(cacheMock, mock.Once()).UpdateViewer(channelName, image, viewerId)
	})
}

func TestRemoveClientWithAnnouncements(t *testing.T) {
	channelName := "channel name"
	viewer := Pet{}

	cacheMock := mock.Mock[ViewerCache]()
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

	announcer.AnnounceJoin(channelName, viewer)
	wg.Wait()

	if len(events) != 0 {
		t.Errorf("expected [] but got %s", events)
	}
}

func TestAnnouncerOnMultipleChannels(t *testing.T) {
	channelOneName := "channel one name"
	channelTwoName := "channel two name"
	viewer := Pet{}

	cacheMock := mock.Mock[ViewerCache]()
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

	announcer.AnnounceJoin(channelOneName, viewer)
	wg.Wait()

	expected := Event{Event: "JOIN", Message: viewer}
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
	viewers := []Pet{{}, {}}

	cacheMock := mock.Mock[ViewerCache]()
	mock.When(cacheMock.GetViewers(channelName)).ThenReturn(viewers)

	announcer := NewAnnouncerService(cacheMock)
	client := announcer.AddClient(channelName)

	got := []Pet{}
	var wg sync.WaitGroup
	wg.Add(len(viewers))

	go func() {
		for event := range client.Stream {
			viewer, ok := event.Message.(Pet)
			if !ok {
				t.Errorf("expected a viewer but got %s", event.Message)
			}
			got = append(got, viewer)
			wg.Done()
		}
	}()

	wg.Wait()
	close(client.Stream)

	if !slices.Equal(got, viewers) {
		t.Errorf("got %s want %s", got, viewers)
	}
}
