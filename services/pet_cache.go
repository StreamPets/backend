package services

import "github.com/streampets/backend/models"

type PetCacheService struct {
	pets map[string]map[models.TwitchId]Pet
}

func NewPetCacheService() *PetCacheService {
	return &PetCacheService{make(map[string]map[models.TwitchId]Pet)}
}

func (s *PetCacheService) AddPet(channelName string, pet Pet) {
	pets, ok := s.pets[channelName]
	if !ok {
		pets = make(map[models.TwitchId]Pet)
		s.pets[channelName] = pets
	}

	pets[pet.UserId] = pet
}

func (s *PetCacheService) RemovePet(channelName string, userId models.TwitchId) {
	pets, ok := s.pets[channelName]
	if !ok {
		return
	}

	delete(pets, userId)
}

func (s *PetCacheService) UpdatePet(channelName, image string, userId models.TwitchId) {
	pets, ok := s.pets[channelName]
	if !ok {
		return
	}

	pet, ok := pets[userId]
	if !ok {
		return
	}

	pet.Image = image
	s.pets[channelName][userId] = pet
}

func (s *PetCacheService) GetPets(channelName string) []Pet {
	pets, ok := s.pets[channelName]
	if !ok {
		return []Pet{}
	}

	result := []Pet{}
	for _, pet := range pets {
		result = append(result, pet)
	}

	return result
}
