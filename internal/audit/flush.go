package audit

import "io"

// Flush writes all entries from a Record to a Logger.
// It returns the number of entries written and any first error encountered.
func Flush(rec *Record, log *Logger) (int, error) {
	if rec == nil || log == nil {
		return 0, nil
	}
	entries := rec.Entries()
	for i, e := range entries {
		if err := log.Write(e); err != nil {
			return i, err
		}
	}
	return len(entries), nil
}

// FlushTo writes all entries from a Record directly to an io.Writer in text format.
func FlushTo(rec *Record, w io.Writer) (int, error) {
	log := NewLogger(w, "text")
	return Flush(rec, log)
}
