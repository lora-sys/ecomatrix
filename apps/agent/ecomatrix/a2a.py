"""A2A v1.1 protocol client.

This module is the Python mirror of apps/backend/pkg/a2a. Keep the envelope
shape, error codes, and timestamp drift window identical to the Go side.

Reference: docs/architecture/api.md
"""

from __future__ import annotations

import re
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Mapping

import httpx

PROTOCOL_V = "1.1"
MAX_CLOCK_SKEW_SECONDS = 300

_MSG_ID_RE = re.compile(r"^[A-Za-z0-9_]{6,64}$")
_AGENT_ID_RE = re.compile(r"^agent_[a-z0-9_]{2,32}$")


class Action(str, Enum):
    EXECUTE_TRADE = "EXECUTE_TRADE"


_ALLOWED_ACTIONS = {a.value for a in Action}


class Currency(str, Enum):
    GOLD = "GOLD"


class Code(str, Enum):
    INVALID_ENVELOPE = "INVALID_ENVELOPE"
    UNKNOWN_ACTION = "UNKNOWN_ACTION"
    PROTOCOL_MISMATCH = "PROTOCOL_MISMATCH"
    UNKNOWN_AGENT = "UNKNOWN_AGENT"
    INSUFFICIENT_FUNDS = "INSUFFICIENT_FUNDS"
    SELF_TRADE = "SELF_TRADE"
    RATE_LIMITED = "RATE_LIMITED"
    INTERNAL = "INTERNAL"
    IDEMPOTENT_REPLAY = "IDEMPOTENT_REPLAY"


class A2AError(Exception):
    def __init__(self, code: Code, message: str, retryable: bool = False, http_status: int | None = None):
        super().__init__(f"a2a:{code.value}:{message}")
        self.code = code
        self.message = message
        self.retryable = retryable
        self.http_status = http_status


@dataclass(frozen=True)
class Offer:
    currency_type: Currency
    amount: int

    def to_dict(self) -> dict[str, Any]:
        return {"currency_type": self.currency_type.value, "amount": self.amount}


@dataclass(frozen=True)
class TradePayload:
    target_agent: str
    offer: Offer
    reasoning: str = ""

    def to_dict(self) -> dict[str, Any]:
        out: dict[str, Any] = {"target_agent": self.target_agent, "offer": self.offer.to_dict()}
        if self.reasoning:
            out["reasoning"] = self.reasoning
        return out


@dataclass(frozen=True)
class Envelope:
    msg_id: str
    sender: str
    action: Action
    payload: Mapping[str, Any]
    timestamp: int = field(default_factory=lambda: int(time.time()))
    protocol_v: str = PROTOCOL_V

    def to_dict(self) -> dict[str, Any]:
        return {
            "protocol_v": self.protocol_v,
            "msg_id": self.msg_id,
            "sender": self.sender,
            "action": self.action.value,
            "payload": dict(self.payload),
            "timestamp": self.timestamp,
        }


@dataclass(frozen=True)
class Receipt:
    tx_id: str
    from_: str
    to: str
    amount: int
    currency_type: str
    balance_after_from: int
    balance_after_to: int

    @classmethod
    def from_dict(cls, d: Mapping[str, Any]) -> "Receipt":
        bal = d.get("balance_after", {})
        return cls(
            tx_id=d["tx_id"],
            from_=d["from"],
            to=d["to"],
            amount=int(d["amount"]),
            currency_type=d["currency_type"],
            balance_after_from=int(bal.get("from", 0)),
            balance_after_to=int(bal.get("to", 0)),
        )


def validate_agent_id(s: str) -> bool:
    return bool(_AGENT_ID_RE.match(s))


def validate_msg_id(s: str) -> bool:
    return bool(_MSG_ID_RE.match(s))


def new_msg_id(prefix: str = "tx") -> str:
    """Generate a unique msg_id satisfying the wire regex."""
    suffix = uuid.uuid4().hex[:12]
    candidate = f"{prefix}_{suffix}"
    return candidate[:64] if len(candidate) > 64 else candidate


def validate_envelope(env: Mapping[str, Any]) -> None:
    """Mirror of pkg/a2a.Validate. Raises A2AError on the first violation."""
    proto = env.get("protocol_v")
    if not proto:
        raise A2AError(Code.INVALID_ENVELOPE, "protocol_v is required")
    if proto != PROTOCOL_V:
        raise A2AError(
            Code.PROTOCOL_MISMATCH,
            f"server speaks {PROTOCOL_V}, got {proto!r}",
        )

    msg_id = env.get("msg_id") or ""
    if not validate_msg_id(msg_id):
        raise A2AError(Code.INVALID_ENVELOPE, f"msg_id {msg_id!r} must match {_MSG_ID_RE.pattern}")

    sender = env.get("sender") or ""
    if not validate_agent_id(sender):
        raise A2AError(Code.INVALID_ENVELOPE, f"sender {sender!r} must match {_AGENT_ID_RE.pattern}")

    action = env.get("action")
    if action not in _ALLOWED_ACTIONS:
        raise A2AError(Code.UNKNOWN_ACTION, f"action {action!r} is not recognized")

    payload = env.get("payload")
    if payload is None:
        raise A2AError(Code.INVALID_ENVELOPE, "payload is required")

    ts = env.get("timestamp")
    if not ts:
        raise A2AError(Code.INVALID_ENVELOPE, "timestamp is required")
    drift = abs(int(time.time()) - int(ts))
    if drift > MAX_CLOCK_SKEW_SECONDS:
        raise A2AError(
            Code.INVALID_ENVELOPE,
            f"timestamp drift {drift}s exceeds {MAX_CLOCK_SKEW_SECONDS}s",
        )


def decode_trade_payload(payload: Mapping[str, Any]) -> TradePayload:
    """Mirror of pkg/a2a.DecodeTradePayload."""
    if payload is None:
        raise A2AError(Code.INVALID_ENVELOPE, "payload is required")
    target = payload.get("target_agent") or ""
    if not validate_agent_id(target):
        raise A2AError(Code.INVALID_ENVELOPE, f"target_agent {target!r} is not a valid agent id")
    offer_raw = payload.get("offer") or {}
    try:
        currency = Currency(str(offer_raw.get("currency_type", "")))
    except ValueError as e:
        raise A2AError(Code.INVALID_ENVELOPE, f"currency_type invalid: {e}") from None
    amount = int(offer_raw.get("amount", 0))
    if amount <= 0:
        raise A2AError(Code.INVALID_ENVELOPE, "offer.amount must be > 0")
    return TradePayload(
        target_agent=target,
        offer=Offer(currency_type=currency, amount=amount),
        reasoning=str(payload.get("reasoning", "")),
    )


class A2AClient:
    """HTTP client for the A2A gateway."""

    def __init__(self, base_url: str, timeout_seconds: float = 5.0) -> None:
        self._base_url = base_url.rstrip("/")
        self._client = httpx.Client(base_url=self._base_url, timeout=timeout_seconds)

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> "A2AClient":
        return self

    def __exit__(self, *exc: Any) -> None:
        self.close()

    # ---- queries ----

    def list_agents(self, limit: int = 50, offset: int = 0) -> list[dict[str, Any]]:
        r = self._client.get("/v1/agents", params={"limit": limit, "offset": offset})
        r.raise_for_status()
        return r.json().get("agents", [])

    def get_agent(self, agent_id: int | str) -> dict[str, Any]:
        r = self._client.get(f"/v1/agents/by-string-id/{agent_id}")
        if r.status_code == 404:
            r = self._client.get(f"/v1/agents/{agent_id}")
        if r.status_code == 404:
            raise A2AError(Code.UNKNOWN_AGENT, f"agent {agent_id} not found", http_status=404)
        r.raise_for_status()
        return r.json()

    def recent_transactions(self, limit: int = 50) -> list[dict[str, Any]]:
        r = self._client.get("/v1/transactions", params={"limit": limit})
        r.raise_for_status()
        return r.json().get("transactions", [])

    # ---- commands ----

    def execute_trade(self, sender: str, target_agent: str, amount: int,
                      reasoning: str = "", msg_id: str | None = None) -> Receipt:
        env = Envelope(
            msg_id=msg_id or new_msg_id("trade"),
            sender=sender,
            action=Action.EXECUTE_TRADE,
            payload=TradePayload(
                target_agent=target_agent,
                offer=Offer(currency_type=Currency.GOLD, amount=amount),
                reasoning=reasoning,
            ).to_dict(),
        )
        r = self._client.post("/v1/trades", json=env.to_dict())
        body = r.json()
        if r.status_code >= 400:
            err = body.get("error", {})
            try:
                code = Code(str(err.get("code", Code.INTERNAL.value)))
            except ValueError:
                code = Code.INTERNAL
            raise A2AError(code, str(err.get("message", "")), retryable=bool(err.get("retryable", False)), http_status=r.status_code)
        if body.get("status") != "SETTLED":
            raise A2AError(Code.INTERNAL, f"unexpected response: {body}", http_status=r.status_code)
        return Receipt.from_dict(body["receipt"])
