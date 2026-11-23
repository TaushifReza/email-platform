package crypto

// LoadMasterKey normalizes a master key supplied through configuration.
// It accepts raw bytes (the production path via Docker secrets) and rejects
// values that are too short to be a 256-bit key. Whitespace is trimmed
// defensively even though the config loader already trims.
func LoadMasterKey(secret []byte) ([]byte, error) {
	out := make([]byte, len(secret))
	copy(out, secret)
	// Trim leading and trailing whitespace, matching config loader behavior.
	start := 0
	for start < len(out) && (out[start] == ' ' || out[start] == '\t' || out[start] == '\n' || out[start] == '\r') {
		start++
	}
	end := len(out)
	for end > start && (out[end-1] == ' ' || out[end-1] == '\t' || out[end-1] == '\n' || out[end-1] == '\r') {
		end--
	}
	out = out[start:end]
	if len(out) < 32 {
		return nil, ErrInvalidMasterKey
	}
	return out, nil
}
