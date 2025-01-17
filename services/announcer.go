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

type PetCache interface {
	AddPet(channelName string, pet Pet)
	RemovePet(channelName string, viewerId models.UserId)
	UpdatePet(channelName, image string, viewerId models.UserId)
	GetPets(channelName string) []Pet
}

type AnnouncerService struct {
	announce      chan wrappedEvent
	newClients    chan Client
	closedClients chan Client
	totalClients  map[string](map[EventStream]bool)
	cache         PetCache
}

func NewAnnouncerService(cache PetCache) *AnnouncerService {
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

func (s *AnnouncerService) AnnounceJoin(channelName string, viewer Pet) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "JOIN",
			Message: viewer,
		},
	}
	s.cache.AddPet(channelName, viewer)
}

func (s *AnnouncerService) AnnouncePart(channelName string, viewerId models.UserId) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "PART",
			Message: viewerId,
		},
	}
	s.cache.RemovePet(channelName, viewerId)
}

func (s *AnnouncerService) AnnounceAction(channelName, action string, viewerId models.UserId) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("%s-%s", action, viewerId),
			Message: viewerId,
		},
	}
}

func (s *AnnouncerService) AnnounceUpdate(channelName, image string, viewerId models.UserId) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("COLOR-%s", viewerId),
			Message: image,
		},
	}
	s.cache.UpdatePet(channelName, image, viewerId)
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

			go func() {
				for _, viewer := range s.cache.GetPets(client.ChannelName) {
					client.Stream <- Event{Event: "JOIN", Message: viewer}
				}
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
