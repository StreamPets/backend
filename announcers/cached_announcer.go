package announcers

import (
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

type announcer interface {
	AddClient(channelId models.TwitchId) Client
	RemoveClient(client Client)
	AnnounceJoin(channelId models.TwitchId, pet services.Pet)
	AnnouncePart(channelId, userId models.TwitchId)
	AnnounceAction(channelId, userId models.TwitchId, action string)
	AnnounceUpdate(channelId, userId models.TwitchId, image string)
}

type CachedAnnouncerService struct {
	announcer announcer
	cache     cacheMap
}

func NewCachedAnnouncerService(
	announcer announcer,
) *CachedAnnouncerService {
	return &CachedAnnouncerService{
		cache:     make(cacheMap),
		announcer: announcer,
	}
}

func (s *CachedAnnouncerService) AddClient(channelId models.TwitchId) Client {
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

func (s *CachedAnnouncerService) RemoveClient(client Client) {
	s.announcer.RemoveClient(client)
}

func (s *CachedAnnouncerService) AnnounceJoin(channelId models.TwitchId, pet services.Pet) {
	pets, ok := s.cache[channelId]
	if !ok {
		pets = make(petMap)
		s.cache[channelId] = pets
	}
	pets[pet.UserId] = pet

	s.announcer.AnnounceJoin(channelId, pet)
}

func (s *CachedAnnouncerService) AnnouncePart(channelId, userId models.TwitchId) {
	pets, ok := s.cache[channelId]
	if !ok {
		return
	}
	delete(pets, userId)

	s.announcer.AnnouncePart(channelId, userId)
}

func (s *CachedAnnouncerService) AnnounceAction(channelId, userId models.TwitchId, action string) {
	s.announcer.AnnounceAction(channelId, userId, action)
}

func (s *CachedAnnouncerService) AnnounceUpdate(channelId, userId models.TwitchId, image string) {
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
