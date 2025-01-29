package services

import "github.com/streampets/backend/models"

type PetCacheService struct {
	pets map[models.TwitchId]map[models.TwitchId]Pet
}

func NewPetCacheService() *PetCacheService {
	return &PetCacheService{make(map[models.TwitchId]map[models.TwitchId]Pet)}
}

func (s *PetCacheService) AddPet(channelId models.TwitchId, pet Pet) {
	pets, ok := s.pets[channelId]
	if !ok {
		pets = make(map[models.TwitchId]Pet)
		s.pets[channelId] = pets
	}

	pets[pet.UserId] = pet
}

func (s *PetCacheService) RemovePet(channelId, userId models.TwitchId) {
	pets, ok := s.pets[channelId]
	if !ok {
		return
	}

	delete(pets, userId)
}

func (s *PetCacheService) UpdatePet(channelId, userId models.TwitchId, image string) {
	pets, ok := s.pets[channelId]
	if !ok {
		return
	}

	pet, ok := pets[userId]
	if !ok {
		return
	}

	pet.Image = image
	s.pets[channelId][userId] = pet
}

func (s *PetCacheService) GetPets(channelId models.TwitchId) []Pet {
	pets, ok := s.pets[channelId]
	if !ok {
		return []Pet{}
	}

	result := []Pet{}
	for _, pet := range pets {
		result = append(result, pet)
	}

	return result
}
