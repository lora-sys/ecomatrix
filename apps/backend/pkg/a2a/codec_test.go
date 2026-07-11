package a2a

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func goodEnvelope() Envelope {
	return Envelope{
		ProtocolV: ProtocolVersion,
		MsgID:     "tx_req_9948",
		Sender:    "agent_miner_01",
		Action:    ActionExecuteTrade,
		Payload: map[string]any{
			"target_agent": "agent_merchant_03",
			"offer": map[string]any{
				"currency_type": "GOLD",
				"amount":        150,
			},
			"reasoning": "vitality low",
		},
		Timestamp: time.Now().Unix(),
	}
}

func TestValidate_HappyPath(t *testing.T) {
	require.NoError(t, Validate(goodEnvelope()))
}

func TestValidate_ProtocolMismatch(t *testing.T) {
	e := goodEnvelope()
	e.ProtocolV = "1.0"
	err := Validate(e)
	require.Error(t, err)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeProtocolMismatch, a2aErr.Code)
}

func TestValidate_MissingMsgID(t *testing.T) {
	e := goodEnvelope()
	e.MsgID = "x"
	err := Validate(e)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestValidate_UnknownAction(t *testing.T) {
	e := goodEnvelope()
	e.Action = "DROP_TABLE"
	err := Validate(e)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeUnknownAction, a2aErr.Code)
}

func TestValidate_ClockSkew(t *testing.T) {
	e := goodEnvelope()
	e.Timestamp = time.Now().Add(-2 * time.Hour).Unix()
	err := Validate(e)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestDecodeTradePayload_SelfTradePassesCodecButBlockedAtService(t *testing.T) {
	// The codec only validates envelope shape; self-trade is caught by the
	// service layer. Here we assert the codec accepts a same-id payload so
	// the service-layer test can own the SELF_TRADE code path.
	e := goodEnvelope()
	e.Sender = "agent_merchant_03"
	e.Payload["target_agent"] = "agent_merchant_03"
	require.NoError(t, Validate(e))
	p, err := DecodeTradePayload(e.Payload)
	require.NoError(t, err)
	assert.Equal(t, e.Sender, p.TargetAgent)
}

func TestDecodeTradePayload_NegativeAmount(t *testing.T) {
	e := goodEnvelope()
	e.Payload["offer"] = map[string]any{"currency_type": "GOLD", "amount": -10}
	_, err := DecodeTradePayload(e.Payload)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestDecodeTradePayload_BadCurrency(t *testing.T) {
	e := goodEnvelope()
	e.Payload["offer"] = map[string]any{"currency_type": "USD", "amount": 100}
	_, err := DecodeTradePayload(e.Payload)
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestDecodeFeedPayload_Happy(t *testing.T) {
	p, err := DecodeFeedPayload(map[string]any{
		"content":     "selling 10 GOLD of iron",
		"intent_type": "OFFER",
	})
	require.NoError(t, err)
	assert.Equal(t, "selling 10 GOLD of iron", p.Content)
	assert.Equal(t, "OFFER", p.IntentType)
}

func TestDecodeFeedPayload_MissingContent(t *testing.T) {
	_, err := DecodeFeedPayload(map[string]any{"intent_type": "OFFER"})
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestDecodeFeedPayload_BadIntentType(t *testing.T) {
	_, err := DecodeFeedPayload(map[string]any{
		"content":     "hi",
		"intent_type": "BULLSHIT",
	})
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}

func TestDecodeFeedPayload_TooLong(t *testing.T) {
	long := make([]byte, 501)
	for i := range long {
		long[i] = 'a'
	}
	_, err := DecodeFeedPayload(map[string]any{
		"content":     string(long),
		"intent_type": "SOCIAL",
	})
	a2aErr, ok := As(err)
	require.True(t, ok)
	assert.Equal(t, CodeInvalidEnvelope, a2aErr.Code)
}
