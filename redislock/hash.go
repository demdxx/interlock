package redislock

import (
	"crypto/sha256"
	"fmt"
)

type objectHash interface {
	ObjectHash() string
}

// Hash value from the object
func hash(obj interface{}) string {
	if ohasher, ok := obj.(objectHash); ok {
		return ohasher.ObjectHash()
	}
	h := sha256.New()
	_, _ = h.Write([]byte(fmt.Sprintf("%v", obj))) // Skip error check
	return fmt.Sprintf("%x", h.Sum(nil))
}
