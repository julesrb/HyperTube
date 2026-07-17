package requestjson

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/respond"
)

const maxJSONBodyBytes = 1 << 20

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		respond.LocalizedError(w, r, http.StatusBadRequest, "BAD_REQUEST", i18n.MsgInvalidJSONBody)
		return false
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respond.LocalizedError(w, r, http.StatusBadRequest, "BAD_REQUEST", i18n.MsgInvalidJSONBody)
		return false
	}

	return true
}

func DecodeJSONObject(w http.ResponseWriter, r *http.Request, allowedFields map[string]struct{}) (map[string]json.RawMessage, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	decoder := json.NewDecoder(r.Body)
	var body map[string]json.RawMessage
	if err := decoder.Decode(&body); err != nil || body == nil {
		respond.LocalizedError(w, r, http.StatusBadRequest, "BAD_REQUEST", i18n.MsgInvalidJSONBody)
		return nil, false
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respond.LocalizedError(w, r, http.StatusBadRequest, "BAD_REQUEST", i18n.MsgInvalidJSONBody)
		return nil, false
	}

	for field := range body {
		if _, ok := allowedFields[field]; !ok {
			respond.LocalizedError(w, r, http.StatusBadRequest, "BAD_REQUEST", i18n.MsgInvalidJSONBody)
			return nil, false
		}
	}

	return body, true
}

func DecodeString(raw json.RawMessage) (string, bool) {
	if IsNull(raw) {
		return "", false
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	return value, true
}

func IsNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
