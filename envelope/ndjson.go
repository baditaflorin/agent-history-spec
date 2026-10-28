package envelope

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ErrLineTooLong ends a read when one line is longer than MaxLineBytes. Lines before it have
// already been handed to the callback.
var ErrLineTooLong = errors.New("envelope: NDJSON line longer than MaxLineBytes")

// ReadNDJSON calls fn once per non-blank line of r with its 1-based line number. A line that is
// not valid JSON for an Event arrives with decodeErr set and a zero Event; the callback decides
// whether to count it as rejected. A non-nil error from fn stops the read and is returned.
// Unknown fields are ignored, so a newer producer does not break an older reader.
func ReadNDJSON(r io.Reader, fn func(line int, e Event, decodeErr error) error) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64<<10), MaxLineBytes)
	n := 0
	for sc.Scan() {
		n++
		b := bytes.TrimSpace(sc.Bytes())
		if len(b) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(b, &e); err != nil {
			e = Event{}
			if cbErr := fn(n, e, err); cbErr != nil {
				return cbErr
			}
			continue
		}
		if err := fn(n, e, nil); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return fmt.Errorf("line %d: %w", n+1, ErrLineTooLong)
		}
		return err
	}
	return nil
}

// WriteNDJSON writes one compact JSON object per event, each ended by a newline, with no HTML escaping.
func WriteNDJSON(w io.Writer, events []Event) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for i := range events {
		if err := enc.Encode(&events[i]); err != nil {
			return fmt.Errorf("event %d: %w", i, err)
		}
	}
	return nil
}
