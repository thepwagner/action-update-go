package updater

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"sort"
)

type SignedUpdateGroup struct {
	Updates   UpdateGroup `json:"signed"`
	Signature []byte      `json:"signature"`
}

func NewSignedUpdateGroup(key []byte, updates UpdateGroup) (SignedUpdateGroup, error) {
	signature, err := updatesHash(key, updates)
	if err != nil {
		return SignedUpdateGroup{}, err
	}
	return SignedUpdateGroup{
		Updates:   updates,
		Signature: signature,
	}, nil
}

func updatesHash(key []byte, updates UpdateGroup) ([]byte, error) {
	sort.Slice(updates.Updates, func(i, j int) bool {
		return updates.Updates[i].Path < updates.Updates[j].Path
	})
	hash := hmac.New(sha512.New, key)
	if err := json.NewEncoder(hash).Encode(&updates); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func VerifySignedUpdateGroup(key []byte, signed SignedUpdateGroup) (*UpdateGroup, error) {
	calculated, err := updatesHash(key, signed.Updates)
	if err != nil {
		return nil, fmt.Errorf("calculating signature: %w", err)
	}
	if subtle.ConstantTimeCompare(calculated, signed.Signature) != 1 {
		return nil, fmt.Errorf("invalid signature")
	}
	return &signed.Updates, nil
}
