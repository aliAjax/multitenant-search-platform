package platform

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewID(prefix string) string {
	b := make([]byte, 8)
	if _, e := rand.Read(b); e != nil {
		return fmt.Sprintf("%s-%d", prefix, 0)
	}
	return prefix + "-" + hex.EncodeToString(b)
}
