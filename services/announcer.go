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

type AnnouncerService struct {
	announce      chan wrappedEvent
	newClients    chan Client
	closedClients chan Client
	totalClients  map[string](map[EventStream]bool)
}

func NewAnnouncerService() *AnnouncerService {
	service := &AnnouncerService{
		announce:      make(chan wrappedEvent),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[string]map[EventStream]bool),
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
}

func (s *AnnouncerService) AnnouncePart(channelName string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   "PART",
			Message: userID,
		},
	}
}

func (s *AnnouncerService) AnnounceAction(channelName, action string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("%s-%s", action, userID),
			Message: userID,
		},
	}
}

func (s *AnnouncerService) AnnounceUpdate(channelName, image string, userID models.TwitchID) {
	s.announce <- wrappedEvent{
		ChannelName: channelName,
		Event: Event{
			Event:   fmt.Sprintf("COLOR-%s", userID),
			Message: image,
		},
	}
}

func (stream *AnnouncerService) listen() {
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
