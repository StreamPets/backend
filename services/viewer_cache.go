package services

import "github.com/streampets/backend/models"

type ViewerCacheService struct {
	viewers map[string]map[models.TwitchId]Pet
}

func NewViewerCacheService() *ViewerCacheService {
	return &ViewerCacheService{make(map[string]map[models.TwitchId]Pet)}
}

func (s *ViewerCacheService) AddViewer(channelName string, viewer Pet) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		viewers = make(map[models.TwitchId]Pet)
		s.viewers[channelName] = viewers
	}

	viewers[viewer.ViewerId] = viewer
}

func (s *ViewerCacheService) RemoveViewer(channelName string, viewerId models.TwitchId) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return
	}

	delete(viewers, viewerId)
}

func (s *ViewerCacheService) UpdateViewer(channelName, image string, viewerId models.TwitchId) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return
	}

	viewer, ok := viewers[viewerId]
	if !ok {
		return
	}

	viewer.Image = image
	s.viewers[channelName][viewerId] = viewer
}

func (s *ViewerCacheService) GetViewers(channelName string) []Pet {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return []Pet{}
	}

	result := []Pet{}
	for _, viewer := range viewers {
		result = append(result, viewer)
	}

	return result
}
