package announcers

import (
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

type AnnouncerService struct {
	announce      chan Announcement
	newClients    chan Client
	closedClients chan Client
	totalClients  map[twitch.Id](map[chan Announcement]bool)
}

func NewAnnouncerService() *AnnouncerService {
	service := &AnnouncerService{
		announce:      make(chan Announcement),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[twitch.Id]map[chan Announcement]bool),
	}

	go service.listen()

	return service
}

func (s *AnnouncerService) AddClient(channelId twitch.Id) Client {
	client := newClient(channelId)
	s.newClients <- client
	return client
}

func (s *AnnouncerService) RemoveClient(client Client) {
	s.closedClients <- client
}

func (s *AnnouncerService) AnnounceJoin(channelId twitch.Id, pet services.Pet) {
	s.announce <- joinAnnouncement(channelId, pet)
}

func (s *AnnouncerService) AnnouncePart(channelId, userId twitch.Id) {
	s.announce <- partAnnouncement(channelId, userId)
}

func (s *AnnouncerService) AnnounceAction(channelId, userId twitch.Id, action string) {
	s.announce <- actionAnnouncement(channelId, userId, action)
}

func (s *AnnouncerService) AnnounceUpdate(channelId, userId twitch.Id, image string) {
	s.announce <- updateAnnouncement(channelId, userId, image)
}

func (s *AnnouncerService) handleNewClient(c Client) {
	_, ok := s.totalClients[c.channelId]
	if !ok {
		s.totalClients[c.channelId] = make(map[chan Announcement]bool)
	}
	s.totalClients[c.channelId][c.Stream] = true
}

func (s *AnnouncerService) handleClosedClient(c Client) {
	delete(s.totalClients[c.channelId], c.Stream)
	close(c.Stream)
}

func (s *AnnouncerService) handleAnnouncement(a Announcement) {
	for eventStream := range s.totalClients[a.channelId] {
		eventStream <- a
	}
}

func (s *AnnouncerService) listen() {
	for {
		select {
		case client := <-s.newClients:
			s.handleNewClient(client)
		case client := <-s.closedClients:
			s.handleClosedClient(client)
		case announcement := <-s.announce:
			s.handleAnnouncement(announcement)
		}
	}
}
