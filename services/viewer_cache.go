package services

import "github.com/streampets/backend/models"

type ViewerCacheService struct {
	viewers map[string]map[models.TwitchID]Viewer
}

func NewViewerCacheService() *ViewerCacheService {
	return &ViewerCacheService{make(map[string]map[models.TwitchID]Viewer)}
}

func (s *ViewerCacheService) AddViewer(channelName string, viewer Viewer) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		viewers = make(map[models.TwitchID]Viewer)
		s.viewers[channelName] = viewers
	}

	viewers[viewer.UserID] = viewer
}

func (s *ViewerCacheService) RemoveViewer(channelName string, viewerID models.TwitchID) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return
	}

	delete(viewers, viewerID)
}

func (s *ViewerCacheService) UpdateViewer(channelName, image string, viewerID models.TwitchID) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return
	}

	viewer, ok := viewers[viewerID]
	if !ok {
		return
	}

	viewer.Image = image
	s.viewers[channelName][viewerID] = viewer
}

func (s *ViewerCacheService) GetViewers(channelName string) []Viewer {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return []Viewer{}
	}

	result := []Viewer{}
	for _, viewer := range viewers {
		result = append(result, viewer)
	}

	return result
}
