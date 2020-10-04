package updater

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"sort"
)

type SignedUpdateDescriptor struct {
	Updates   []Update `json:"updates"`
	Signature []byte   `json:"signature"`
}

func NewSignedUpdateDescriptor(key []byte, updates ...Update) (SignedUpdateDescriptor, error) {
	signature, err := updatesHash(key, updates)
	if err != nil {
		return SignedUpdateDescriptor{}, err
	}
	return SignedUpdateDescriptor{
		Updates:   updates,
		Signature: signature,
	}, nil
}

func updatesHash(key []byte, updates []Update) ([]byte, error) {
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].Path < updates[j].Path
	})
	hash := hmac.New(sha512.New, key)
	if err := json.NewEncoder(hash).Encode(updates); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func VerifySignedUpdateDescriptor(key []byte, descriptor SignedUpdateDescriptor) ([]Update, error) {
	calculated, err := updatesHash(key, descriptor.Updates)
	if err != nil {
		return nil, fmt.Errorf("calculating signature: %w", err)
	}
	if subtle.ConstantTimeCompare(calculated, descriptor.Signature) != 1 {
		return nil, fmt.Errorf("invalid signature")
	}
	return descriptor.Updates, nil
}
