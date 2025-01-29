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
	ChannelId models.TwitchId
	Event     Event
}

type Client struct {
	ChannelId models.TwitchId
	Stream    EventStream
}

type PetCache interface {
	AddPet(channelId models.TwitchId, pet Pet)
	RemovePet(channelId, userId models.TwitchId)
	UpdatePet(channelId, userId models.TwitchId, image string)
	GetPets(channelId models.TwitchId) []Pet
}

type AnnouncerService struct {
	announce      chan wrappedEvent
	newClients    chan Client
	closedClients chan Client
	totalClients  map[models.TwitchId](map[EventStream]bool)
	cache         PetCache
}

func NewAnnouncerService(cache PetCache) *AnnouncerService {
	service := &AnnouncerService{
		announce:      make(chan wrappedEvent),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[models.TwitchId]map[EventStream]bool),
		cache:         cache,
	}

	go service.listen()

	return service
}

func (s *AnnouncerService) AddClient(channelId models.TwitchId) Client {
	client := Client{ChannelId: channelId, Stream: make(EventStream)}
	s.newClients <- client
	return client
}

func (s *AnnouncerService) RemoveClient(client Client) {
	s.closedClients <- client
}

func (s *AnnouncerService) AnnounceJoin(channelId models.TwitchId, pet Pet) {
	s.announce <- wrappedEvent{
		ChannelId: channelId,
		Event: Event{
			Event:   "JOIN",
			Message: pet,
		},
	}
	s.cache.AddPet(channelId, pet)
}

func (s *AnnouncerService) AnnouncePart(channelId, userId models.TwitchId) {
	s.announce <- wrappedEvent{
		ChannelId: channelId,
		Event: Event{
			Event:   "PART",
			Message: userId,
		},
	}
	s.cache.RemovePet(channelId, userId)
}

func (s *AnnouncerService) AnnounceAction(channelId, userId models.TwitchId, action string) {
	s.announce <- wrappedEvent{
		ChannelId: channelId,
		Event: Event{
			Event:   fmt.Sprintf("%s-%s", action, userId),
			Message: userId,
		},
	}
}

func (s *AnnouncerService) AnnounceUpdate(channelId, userId models.TwitchId, image string) {
	s.announce <- wrappedEvent{
		ChannelId: channelId,
		Event: Event{
			Event:   fmt.Sprintf("COLOR-%s", userId),
			Message: image,
		},
	}
	s.cache.UpdatePet(channelId, userId, image)
}

func (s *AnnouncerService) listen() {
	for {
		select {

		case client := <-s.newClients:
			_, ok := s.totalClients[client.ChannelId]
			if !ok {
				s.totalClients[client.ChannelId] = make(map[EventStream]bool)
			}
			s.totalClients[client.ChannelId][client.Stream] = true

			go func() {
				for _, pet := range s.cache.GetPets(client.ChannelId) {
					client.Stream <- Event{Event: "JOIN", Message: pet}
				}
			}()

		case client := <-s.closedClients:
			delete(s.totalClients[client.ChannelId], client.Stream)
			close(client.Stream)

		case wrappedEvent := <-s.announce:
			for eventStream := range s.totalClients[wrappedEvent.ChannelId] {
				eventStream <- wrappedEvent.Event
			}
		}
	}
}
