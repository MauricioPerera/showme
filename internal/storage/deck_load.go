package storage

import (
	"encoding/json"
	"os"

	"github.com/MauricioPerera/showme/internal/domain"
)

// LoadDeck reads the JSON file at path and deserializes it into a domain.Deck.
func LoadDeck(path string) (domain.Deck, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.Deck{}, err
	}

	var deck domain.Deck
	if err := json.Unmarshal(data, &deck); err != nil {
		return domain.Deck{}, err
	}

	return deck, nil
}
