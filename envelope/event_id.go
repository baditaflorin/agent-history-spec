package envelope

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"
)

// ID returns the event id. A producer that has a native id which is unique across all
// sessions (a Claude message uuid, a tool_use id) passes it as native and gets it back as is.
// With no native id the id is HashID.
func ID(native, sessionKey string, line int64, content string) string {
	if native != "" {
		return native
	}
	return HashID(sessionKey, line, content)
}

// HashID is "sha256:" plus the hex sha256 of session_key, NUL, the decimal line number, NUL
// and the content. The line is the source's own position (a line in a JSONL file, a message
// index in a snapshot), so equal content on two lines gets two ids.
func HashID(sessionKey string, line int64, content string) string {
	h := sha256.New()
	_, _ = io.WriteString(h, sessionKey)
	_, _ = h.Write([]byte{0})
	_, _ = io.WriteString(h, strconv.FormatInt(line, 10))
	_, _ = h.Write([]byte{0})
	_, _ = io.WriteString(h, content)
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// ScopedID is for a native id that is only unique inside one session, such as a Codex
// call_id: it prefixes the session key so two sessions can never collide. A native id that
// is already global must not be scoped, so that a resumed or copied transcript collapses
// into the rows already stored.
func ScopedID(sessionKey, native string) string {
	return sessionKey + "#" + native
}
