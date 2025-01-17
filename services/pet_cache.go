package services

import "github.com/streampets/backend/models"

type PetCacheService struct {
	viewers map[string]map[models.UserId]Pet
}

func NewPetCacheService() *PetCacheService {
	return &PetCacheService{make(map[string]map[models.UserId]Pet)}
}

func (s *PetCacheService) AddPet(channelName string, pet Pet) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		viewers = make(map[models.UserId]Pet)
		s.viewers[channelName] = viewers
	}

	viewers[pet.ViewerId] = pet
}

func (s *PetCacheService) RemovePet(channelName string, viewerId models.UserId) {
	viewers, ok := s.viewers[channelName]
	if !ok {
		return
	}

	delete(viewers, viewerId)
}

func (s *PetCacheService) UpdatePet(channelName, image string, viewerId models.UserId) {
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

func (s *PetCacheService) GetPets(channelName string) []Pet {
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
