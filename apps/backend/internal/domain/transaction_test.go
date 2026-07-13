package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReceipt_JSONShape_MatchesA2AProtocol(t *testing.T) {
	receipt := Receipt{
		TxID:         "tx_contract_01",
		From:         "agent_miner_01",
		To:           "agent_merchant_01",
		Amount:       7,
		CurrencyType: "GOLD",
	}
	receipt.BalanceAfter.From = 93
	receipt.BalanceAfter.To = 207

	raw, err := json.Marshal(receipt)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	require.Equal(t, "tx_contract_01", body["tx_id"])
	require.Equal(t, "agent_miner_01", body["from"])
	require.Equal(t, "agent_merchant_01", body["to"])
	require.Equal(t, "GOLD", body["currency_type"])
	require.NotContains(t, body, "TxID")
	require.Equal(t, map[string]any{
		"from": float64(93),
		"to":   float64(207),
	}, body["balance_after"])
}
