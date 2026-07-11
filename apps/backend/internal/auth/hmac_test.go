package auth

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeSignature_Deterministic(t *testing.T) {
	secret := []byte("super-secret")
	ts := int64(1713532588)
	body := []byte(`{"protocol_v":"1.1","msg_id":"tx_req_9948"}`)
	a := ComputeSignature(secret, "POST", "/v1/trades", ts, body)
	b := ComputeSignature(secret, "POST", "/v1/trades", ts, body)
	assert.Equal(t, a, b)
	assert.Len(t, a, 64) // hex(sha256) = 64 chars
}

func TestVerify_HappyPath(t *testing.T) {
	secret := []byte("super-secret")
	body := []byte(`{"x":"y"}`)
	ts := time.Now().Unix()
	hdrs := FormatSignatureHeaders(secret, "agent_miner_01", "POST", "/v1/trades", ts, body)
	require.NoError(t, Verify(secret,
		hdrs[HeaderAgentId],
		hdrs[HeaderAgentTimestamp],
		hdrs[HeaderAgentSignature],
		"POST", "/v1/trades", body,
	))
}

func TestVerify_ExpiredTimestamp(t *testing.T) {
	secret := []byte("super-secret")
	body := []byte(`{}`)
	old := time.Now().Add(-2 * MaxClockSkew).Unix()
	hdrs := FormatSignatureHeaders(secret, "agent_miner_01", "POST", "/v1/trades", old, body)
	err := Verify(secret,
		hdrs[HeaderAgentId],
		hdrs[HeaderAgentTimestamp],
		hdrs[HeaderAgentSignature],
		"POST", "/v1/trades", body,
	)
	assert.ErrorIs(t, err, ErrTimestampSkew)
}

func TestVerify_TamperedBody(t *testing.T) {
	secret := []byte("super-secret")
	ts := time.Now().Unix()
	hdrs := FormatSignatureHeaders(secret, "agent_miner_01", "POST", "/v1/trades", ts, []byte(`{"x":"y"}`))
	tampered := []byte(`{"x":"z"}`)
	err := Verify(secret,
		hdrs[HeaderAgentId],
		hdrs[HeaderAgentTimestamp],
		hdrs[HeaderAgentSignature],
		"POST", "/v1/trades", tampered,
	)
	assert.ErrorIs(t, err, ErrSignatureMismatch)
}

func TestVerify_WrongSecret(t *testing.T) {
	ts := time.Now().Unix()
	body := []byte(`{}`)
	signed := FormatSignatureHeaders([]byte("secret-a"), "agent_miner_01", "POST", "/v1/trades", ts, body)
	err := Verify([]byte("secret-b"),
		signed[HeaderAgentId],
		signed[HeaderAgentTimestamp],
		signed[HeaderAgentSignature],
		"POST", "/v1/trades", body,
	)
	assert.ErrorIs(t, err, ErrSignatureMismatch)
}

func TestVerify_MissingHeaders(t *testing.T) {
	err := Verify([]byte("k"), "", "", "", "POST", "/v1/trades", nil)
	assert.ErrorIs(t, err, ErrMissingHeaders)

	err = Verify([]byte("k"), "agent_x", "1", "", "POST", "/v1/trades", nil)
	assert.ErrorIs(t, err, ErrMissingHeaders)
}

func TestVerify_BadTimestamp(t *testing.T) {
	err := Verify([]byte("k"), "agent_x", "not-a-number", "deadbeef", "POST", "/v1/trades", nil)
	assert.ErrorIs(t, err, ErrTimestampParse)
}

func TestVerify_FutureTimestampBeyondSkew(t *testing.T) {
	body := []byte(`{}`)
	future := time.Now().Add(2 * MaxClockSkew).Unix()
	hdrs := FormatSignatureHeaders([]byte("k"), "agent_x", "POST", "/v1/trades", future, body)
	err := Verify([]byte("k"),
		hdrs[HeaderAgentId],
		hdrs[HeaderAgentTimestamp],
		hdrs[HeaderAgentSignature],
		"POST", "/v1/trades", body,
	)
	assert.ErrorIs(t, err, ErrTimestampSkew)
}

func TestVerify_ConstantTime(t *testing.T) {
	// Smoke test: verify that two equal signatures pass. (We can't easily
	// test for constant-time, but we exercise the hmac.Equal path.)
	body := []byte(`{"k":"v"}`)
	ts := time.Now().Unix()
	hdrs := FormatSignatureHeaders([]byte("k"), "agent_x", "POST", "/v1/trades", ts, body)
	err := Verify([]byte("k"),
		hdrs[HeaderAgentId],
		hdrs[HeaderAgentTimestamp],
		hdrs[HeaderAgentSignature],
		"POST", "/v1/trades", body,
	)
	require.NoError(t, err)
	_ = strconv.Itoa // keep import
}
