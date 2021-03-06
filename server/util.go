// Copyright 2013 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

package server

import (
	"bytes"
	"log"
	"strconv"
	"time"
)

func bencode(data interface{}, buf *bytes.Buffer) {
	switch v := data.(type) {
	case string:
		buf.WriteString(strconv.Itoa(len(v)))
		buf.WriteRune(':')
		buf.WriteString(v)
	case int:
		buf.WriteRune('i')
		buf.WriteString(strconv.Itoa(v))
		buf.WriteRune('e')
	case uint:
		buf.WriteRune('i')
		buf.WriteString(strconv.FormatUint(uint64(v), 10))
		buf.WriteRune('e')
	case int64:
		buf.WriteRune('i')
		buf.WriteString(strconv.FormatInt(v, 10))
		buf.WriteRune('e')
	case uint64:
		buf.WriteRune('i')
		buf.WriteString(strconv.FormatUint(v, 10))
		buf.WriteRune('e')
	case time.Duration:
		// Assume seconds
		buf.WriteRune('i')
		buf.WriteString(strconv.FormatInt(int64(v/time.Second), 10))
		buf.WriteRune('e')
	case map[string]interface{}:
		buf.WriteRune('d')
		for key, val := range v {
			buf.WriteString(strconv.Itoa(len(key)))
			buf.WriteRune(':')
			buf.WriteString(key)
			bencode(val, buf)
		}
		buf.WriteRune('e')
	case []string:
		buf.WriteRune('l')
		for _, val := range v {
			bencode(val, buf)
		}
		buf.WriteRune('e')
	default:
		// Should handle []interface{} manually since Go can't do it implicitly (not currently necessary though)
		log.Printf("%T\n", v)
		panic("Tried to bencode an unsupported type!")
	}
}

// MinInt returns the smaller of the two integers provided.
func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
