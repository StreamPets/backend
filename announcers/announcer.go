package announcers

import (
	"github.com/streampets/backend/pets"
)

type Announcer struct {
	announce      chan Announcement
	newClients    chan Client
	closedClients chan Client
	totalClients  map[string](map[chan Announcement]bool)
}

func NewAnnouncer() *Announcer {
	service := &Announcer{
		announce:      make(chan Announcement),
		newClients:    make(chan Client),
		closedClients: make(chan Client),
		totalClients:  make(map[string]map[chan Announcement]bool),
	}

	go service.listen()

	return service
}

func (s *Announcer) AddClient(channelId string) Client {
	client := newClient(channelId)
	s.newClients <- client
	return client
}

func (s *Announcer) RemoveClient(client Client) {
	s.closedClients <- client
}

func (s *Announcer) AnnounceJoin(channelId string, pet pets.Pet) {
	s.announce <- joinAnnouncement(channelId, pet)
}

func (s *Announcer) AnnouncePart(channelId, userId string) {
	s.announce <- partAnnouncement(channelId, userId)
}

func (s *Announcer) AnnounceAction(channelId, userId string, action string) {
	s.announce <- actionAnnouncement(channelId, userId, action)
}

func (s *Announcer) AnnounceUpdate(channelId, userId string, image string) {
	s.announce <- updateAnnouncement(channelId, userId, image)
}

func (s *Announcer) handleNewClient(c Client) {
	_, ok := s.totalClients[c.channelId]
	if !ok {
		s.totalClients[c.channelId] = make(map[chan Announcement]bool)
	}
	s.totalClients[c.channelId][c.Stream] = true
}

func (s *Announcer) handleClosedClient(c Client) {
	delete(s.totalClients[c.channelId], c.Stream)
	close(c.Stream)
}

func (s *Announcer) handleAnnouncement(a Announcement) {
	for eventStream := range s.totalClients[a.channelId] {
		eventStream <- a
	}
}

func (s *Announcer) listen() {
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
