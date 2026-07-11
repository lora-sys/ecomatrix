// Package auth implements per-agent HMAC signing for A2A endpoints.
//
// The scheme mirrors the A2A envelope:
//
//   string_to_sign = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + sha256_hex(BODY)
//   signature     = hex_hmac_sha256(secret, string_to_sign)
//
// Headers expected on every A2A request (besides Content-Type):
//
//   X-Agent-Id:        agent_miner_01
//   X-Agent-Timestamp: 1713532588          (unix seconds)
//   X-Agent-Signature: <hex digest>
//
// The signed timestamp is also checked against MaxClockSkew (5 minutes) to
// prevent replay. Admin endpoints (POST /v1/agents) still require the shared
// X-Admin-Token; this scheme authenticates *agents*, not operators.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

// MaxClockSkew bounds how old a signed request may be.
const MaxClockSkew = 5 * time.Minute

// Header keys.
const (
	HeaderAgentId        = "X-Agent-Id"
	HeaderAgentTimestamp = "X-Agent-Timestamp"
	HeaderAgentSignature = "X-Agent-Signature"
)

// Sentinel errors.
var (
	ErrMissingHeaders    = errors.New("auth: missing agent signature headers")
	ErrTimestampParse    = errors.New("auth: invalid X-Agent-Timestamp")
	ErrTimestampSkew     = errors.New("auth: timestamp out of window")
	ErrSignatureMismatch = errors.New("auth: signature mismatch")
)

// ComputeSignature produces the signature for a given method/path/body/timestamp
// using the provided shared secret.
func ComputeSignature(secret []byte, method, path string, ts int64, body []byte) string {
	h := hmac.New(sha256.New, secret)
	bodyHash := sha256.Sum256(body)
	canonical := strings.Join([]string{
		method,
		path,
		strconv.FormatInt(ts, 10),
		hex.EncodeToString(bodyHash[:]),
	}, "\n")
	h.Write([]byte(canonical))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify checks the signature headers against the secret for the named agent.
// Returns nil on success; a sentinel error otherwise.
func Verify(secret []byte, agentID, headerTS, headerSig, method, path string, body []byte) error {
	if headerTS == "" || headerSig == "" || agentID == "" {
		return ErrMissingHeaders
	}
	ts, err := strconv.ParseInt(headerTS, 10, 64)
	if err != nil {
		return ErrTimestampParse
	}
	drift := time.Since(time.Unix(ts, 0))
	if drift < -MaxClockSkew || drift > MaxClockSkew {
		return ErrTimestampSkew
	}
	expected := ComputeSignature(secret, method, path, ts, body)
	if !hmac.Equal([]byte(expected), []byte(headerSig)) {
		return ErrSignatureMismatch
	}
	return nil
}

// FormatSignatureHeaders returns the three headers for a client to set.
// Convenience helper for the Python side and for tests.
func FormatSignatureHeaders(secret []byte, agentID, method, path string, ts int64, body []byte) map[string]string {
	return map[string]string{
		HeaderAgentId:        agentID,
		HeaderAgentTimestamp: strconv.FormatInt(ts, 10),
		HeaderAgentSignature: ComputeSignature(secret, method, path, ts, body),
	}
}
