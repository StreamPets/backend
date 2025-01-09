package services

import "github.com/streampets/backend/models"

type ViewerCacheService struct {
	viewers map[models.TwitchID]map[models.TwitchID]Viewer
}

func NewViewerCacheService() *ViewerCacheService {
	return &ViewerCacheService{make(map[models.TwitchID]map[models.TwitchID]Viewer)}
}

func (s *ViewerCacheService) AddViewer(channelID models.TwitchID, viewer Viewer) {
	viewers, ok := s.viewers[channelID]
	if !ok {
		viewers = make(map[models.TwitchID]Viewer)
		s.viewers[channelID] = viewers
	}

	viewers[viewer.UserID] = viewer
}

func (s *ViewerCacheService) RemoveViewer(channelID, viewerID models.TwitchID) {
	viewers, ok := s.viewers[channelID]
	if !ok {
		return
	}

	delete(viewers, viewerID)
}

func (s *ViewerCacheService) UpdateViewer(channelID, viewerID models.TwitchID, image string) {
	viewers, ok := s.viewers[channelID]
	if !ok {
		return
	}

	viewer, ok := viewers[viewerID]
	if !ok {
		return
	}

	viewer.Image = image
	s.viewers[channelID][viewerID] = viewer
}

func (s *ViewerCacheService) GetViewers(channelID models.TwitchID) []Viewer {
	viewers, ok := s.viewers[channelID]
	if !ok {
		return []Viewer{}
	}

	result := []Viewer{}
	for _, viewer := range viewers {
		result = append(result, viewer)
	}

	return result
}
