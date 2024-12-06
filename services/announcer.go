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

// TODO: Come up with a better name
type wrappedEvent struct {
	ChannelName string
	Event       Event
}

type Client struct {
	ChannelName string
	Stream      EventStream
}

type Announcer interface {
	AddClient(channelName string) Client
	RemoveClient(client Client)

	AnnounceJoin(channelName string, viewer Viewer)
	AnnouncePart(channelName string, userID models.TwitchID)
	AnnounceAction(channelName, action string, userID models.TwitchID)
	AnnounceUpdate(channelName, image string, userID models.TwitchID)
}

type announceService struct {
	announce      chan wrappedEvent
	newClients    chan Client
	closedClients chan Client
	totalClients  map[string](map[EventStream]bool)
}

func NewAnnounceService() Announcer {
	service := &announceService{
		announce:      make(chan wrappedEvent),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[string]map[EventStream]bool),
	}

	go service.listen()

	return service
}

func (s *announceService) AddClient(channelName string) Client {
	client := Client{ChannelName: channelName, Stream: make(EventStream)}
	s.newClients <- client
	return client
}

func (s *announceService) RemoveClient(client Client) {
	s.closedClients <- client
}

func (s *announceService) AnnounceJoin(channelName string, viewer Viewer) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "JOIN",
			Message: viewer,
		},
	}
}

func (s *announceService) AnnouncePart(channelName string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "PART",
			Message: userID,
		},
	}
}

func (s *announceService) AnnounceAction(channelName, action string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("%s-%s", action, userID),
			Message: userID,
		},
	}
}

func (s *announceService) AnnounceUpdate(channelName, image string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event: "ITEM",
			Message: map[string]string{
				"image":  image,
				"userID": string(userID),
			},
		},
	}
}

func (stream *announceService) listen() {
	for {
		select {

		case client := <-stream.newClients:
			_, ok := stream.totalClients[client.ChannelName]
			if !ok {
				stream.totalClients[client.ChannelName] = make(map[EventStream]bool)
			}
			stream.totalClients[client.ChannelName][client.Stream] = true

		case client := <-stream.closedClients:
			delete(stream.totalClients[client.ChannelName], client.Stream)
			close(client.Stream)

		case wrappedEvent := <-stream.announce:
			for eventStream := range stream.totalClients[wrappedEvent.ChannelName] {
				eventStream <- wrappedEvent.Event
			}
		}
	}
}
