package query

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Cursor struct {
	Offset    int    `json:"offset"`
	SortValue string `json:"sort_value"`
	Issued    int64  `json:"issued"`
}

func EncodeCursor(c Cursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}
func DecodeCursor(s string) (Cursor, error) {
	b, e := base64.RawURLEncoding.DecodeString(s)
	if e != nil {
		return Cursor{}, fmt.Errorf("decode cursor: %w", e)
	}
	var c Cursor
	if e = json.Unmarshal(b, &c); e != nil {
		return c, fmt.Errorf("parse cursor: %w", e)
	}
	return c, nil
}
