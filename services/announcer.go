package services

import (
	"fmt"

	"github.com/streampets/backend/models"
)

type Event struct {
	Event   string
	Message interface{}
}

type EventStream chan Event

type wrappedEvent struct {
	ChannelName string
	Event       Event
}

type Client struct {
	ChannelName string
	Stream      EventStream
}

type ViewerCache interface {
	AddViewer(channelName string, viewer Viewer)
	RemoveViewer(channelName string, viewerID models.TwitchID)
	UpdateViewer(channelName, image string, viewerID models.TwitchID)
	GetViewers(channelName string) []Viewer
}

type AnnouncerService struct {
	announce      chan wrappedEvent
	newClients    chan Client
	closedClients chan Client
	totalClients  map[string](map[EventStream]bool)
	cache         ViewerCache
}

func NewAnnouncerService(cache ViewerCache) *AnnouncerService {
	service := &AnnouncerService{
		announce:      make(chan wrappedEvent),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[string]map[EventStream]bool),
		cache:         cache,
	}

	go service.listen()

	return service
}

func (s *AnnouncerService) AddClient(channelName string) Client {
	client := Client{ChannelName: channelName, Stream: make(EventStream)}
	s.newClients <- client
	return client
}

func (s *AnnouncerService) RemoveClient(client Client) {
	s.closedClients <- client
}

func (s *AnnouncerService) AnnounceJoin(channelName string, viewer Viewer) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "JOIN",
			Message: viewer,
		},
	}
	s.cache.AddViewer(channelName, viewer)
}

func (s *AnnouncerService) AnnouncePart(channelName string, viewerID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "PART",
			Message: viewerID,
		},
	}
	s.cache.RemoveViewer(channelName, viewerID)
}

func (s *AnnouncerService) AnnounceAction(channelName, action string, viewerID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("%s-%s", action, viewerID),
			Message: viewerID,
		},
	}
}

func (s *AnnouncerService) AnnounceUpdate(channelName, image string, viewerID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("COLOR-%s", viewerID),
			Message: image,
		},
	}
	s.cache.UpdateViewer(channelName, image, viewerID)
}

func (s *AnnouncerService) listen() {
	for {
		select {

		case client := <-s.newClients:
			_, ok := s.totalClients[client.ChannelName]
			if !ok {
				s.totalClients[client.ChannelName] = make(map[EventStream]bool)
			}
			s.totalClients[client.ChannelName][client.Stream] = true

			// Initialise a WG here
			go func() {
				for _, viewer := range s.cache.GetViewers(client.ChannelName) {
					client.Stream <- Event{Event: "JOIN", Message: viewer}
				}
				// Call WG.Done() here
			}()

		case client := <-s.closedClients:
			delete(s.totalClients[client.ChannelName], client.Stream)
			close(client.Stream)

		case wrappedEvent := <-s.announce:
			for eventStream := range s.totalClients[wrappedEvent.ChannelName] {
				eventStream <- wrappedEvent.Event
			}
		}
	}
}
