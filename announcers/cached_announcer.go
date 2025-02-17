package announcers

import (
	"github.com/streampets/backend/pets"
)

type announcer interface {
	AddClient(channelId string) Client
	RemoveClient(client Client)
	AnnounceJoin(channelId string, pet pets.Pet)
	AnnouncePart(channelId, userId string)
	AnnounceAction(channelId, userId string, action string)
	AnnounceUpdate(channelId, userId string, image string)
}

type CachedAnnouncer struct {
	announcer announcer
	cache     cacheMap
}

func NewCachedAnnouncer(
	announcer announcer,
) *CachedAnnouncer {
	return &CachedAnnouncer{
		cache:     make(cacheMap),
		announcer: announcer,
	}
}

func (s *CachedAnnouncer) AddClient(channelId string) Client {
	client := s.announcer.AddClient(channelId)

	go func() {
		pets, ok := s.cache[channelId]
		if ok {
			for _, pet := range pets {
				client.Stream <- joinAnnouncement(channelId, pet)
			}
		}
	}()

	return client
}

func (s *CachedAnnouncer) RemoveClient(client Client) {
	s.announcer.RemoveClient(client)
}

func (s *CachedAnnouncer) AnnounceJoin(channelId string, pet pets.Pet) {
	pets, ok := s.cache[channelId]
	if !ok {
		pets = make(petMap)
		s.cache[channelId] = pets
	}
	pets[pet.UserId] = pet

	s.announcer.AnnounceJoin(channelId, pet)
}

func (s *CachedAnnouncer) AnnouncePart(channelId, userId string) {
	pets, ok := s.cache[channelId]
	if !ok {
		return
	}
	delete(pets, userId)

	s.announcer.AnnouncePart(channelId, userId)
}

func (s *CachedAnnouncer) AnnounceAction(channelId, userId string, action string) {
	s.announcer.AnnounceAction(channelId, userId, action)
}

func (s *CachedAnnouncer) AnnounceUpdate(channelId, userId string, image string) {
	pets, ok := s.cache[channelId]
	if !ok {
		return
	}

	pet, ok := pets[userId]
	if !ok {
		return
	}

	pet.Image = image
	s.cache[channelId][userId] = pet

	s.announcer.AnnounceUpdate(channelId, userId, image)
}
