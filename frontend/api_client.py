"""API client for TinyStock Go backend with JWT auth support."""

import os
from typing import Any

import requests

API_BASE = os.getenv("TINYSTOCK_API_URL", "http://localhost:8080")


def _url(path: str) -> str:
    return f"{API_BASE.rstrip('/')}{path}"


def _headers(token: str | None = None) -> dict:
    h = {"Content-Type": "application/json"}
    if token:
        h["Authorization"] = f"Bearer {token}"
    return h


def _get_data(r: requests.Response) -> Any:
    """Extract data from API response."""
    try:
        data = r.json()
        if "data" in data:
            return data["data"]
        return data
    except Exception:
        return None


def _get_error(r: requests.Response) -> str:
    """Extract error message from API response."""
    try:
        data = r.json()
        if "error" in data and isinstance(data["error"], dict):
            return data["error"].get("message", str(data["error"]))
        return r.text or "Request failed"
    except Exception:
        return r.text or "Request failed"


# --- Auth ---

def login(email: str, password: str) -> tuple[bool, str | None, dict | None]:
    """Login and return (success, token, user)."""
    try:
        r = requests.post(
            _url("/api/auth/login"),
            json={"email": email, "password": password},
            headers=_headers(),
            timeout=10,
        )
        data = _get_data(r)
        if r.status_code == 200 and data and "token" in data:
            return True, data["token"], data.get("user")
        return False, None, None
    except requests.RequestException as e:
        return False, None, None


def register(email: str, password: str) -> tuple[bool, str]:
    """Register new user. Returns (success, message)."""
    try:
        r = requests.post(
            _url("/api/auth/register"),
            json={"email": email, "password": password},
            headers=_headers(),
            timeout=10,
        )
        if r.status_code == 201:
            return True, "Registration successful"
        return False, _get_error(r)
    except requests.RequestException as e:
        return False, str(e)


# --- Stock (public, no auth) ---

def get_quote(symbol: str) -> dict[str, Any] | None:
    """Fetch current quote for a symbol."""
    try:
        r = requests.get(_url(f"/api/quote/{symbol}"), timeout=10)
        r.raise_for_status()
        return _get_data(r)
    except requests.RequestException:
        return None


def get_history(symbol: str, range_: str = "1mo", interval: str = "1d") -> list[dict] | None:
    """Fetch historical price data."""
    try:
        r = requests.get(
            _url(f"/api/history/{symbol}"),
            params={"range": range_, "interval": interval},
            timeout=10,
        )
        r.raise_for_status()
        data = _get_data(r)
        return data.get("history", []) if isinstance(data, dict) else []
    except requests.RequestException:
        return None


def search_symbols(query: str, limit: int = 10) -> list[dict]:
    """Search for stock symbols."""
    try:
        r = requests.get(
            _url("/api/search"),
            params={"q": query, "limit": limit},
            timeout=10,
        )
        r.raise_for_status()
        data = _get_data(r)
        return data.get("results", []) if isinstance(data, dict) else []
    except requests.RequestException:
        return []


# --- Watchlist (requires auth) ---

def get_watchlist(token: str) -> tuple[list[dict], list[dict]]:
    """Get watchlist with current quotes."""
    try:
        r = requests.get(_url("/api/watchlist"), headers=_headers(token), timeout=10)
        r.raise_for_status()
        data = _get_data(r)
        if isinstance(data, dict):
            return data.get("watchlist", []), data.get("quotes", [])
        return [], []
    except requests.RequestException:
        return [], []


def add_to_watchlist(token: str, symbol: str) -> tuple[bool, str]:
    """Add symbol to watchlist."""
    try:
        r = requests.post(
            _url("/api/watchlist"),
            json={"symbol": symbol},
            headers=_headers(token),
            timeout=10,
        )
        if r.status_code == 201:
            return True, "Added to watchlist"
        return False, _get_error(r)
    except requests.RequestException as e:
        return False, str(e)


def remove_from_watchlist(token: str, symbol: str) -> tuple[bool, str]:
    """Remove symbol from watchlist."""
    try:
        r = requests.delete(
            _url(f"/api/watchlist/{symbol}"),
            headers=_headers(token),
            timeout=10,
        )
        if r.status_code == 200:
            return True, "Removed from watchlist"
        return False, _get_error(r)
    except requests.RequestException as e:
        return False, str(e)


# --- Portfolio (requires auth) ---

def get_portfolio(token: str) -> dict[str, Any]:
    """Get portfolio holdings with P&L."""
    try:
        r = requests.get(_url("/api/portfolio"), headers=_headers(token), timeout=10)
        r.raise_for_status()
        data = _get_data(r)
        if isinstance(data, dict):
            return data
        return {"holdings": [], "totalValue": 0, "totalCost": 0, "totalPnL": 0, "returnPct": 0}
    except requests.RequestException:
        return {"holdings": [], "totalValue": 0, "totalCost": 0, "totalPnL": 0, "returnPct": 0}


def add_holding(token: str, symbol: str, quantity: float, buy_price: float) -> tuple[bool, str]:
    """Add holding to portfolio."""
    try:
        r = requests.post(
            _url("/api/portfolio"),
            json={"symbol": symbol, "quantity": quantity, "buyPrice": buy_price},
            headers=_headers(token),
            timeout=10,
        )
        if r.status_code == 201:
            return True, "Added to portfolio"
        return False, _get_error(r)
    except requests.RequestException as e:
        return False, str(e)


def remove_holding(token: str, holding_id: int) -> tuple[bool, str]:
    """Remove holding from portfolio."""
    try:
        r = requests.delete(
            _url(f"/api/portfolio/{holding_id}"),
            headers=_headers(token),
            timeout=10,
        )
        if r.status_code == 200:
            return True, "Removed from portfolio"
        return False, _get_error(r)
    except requests.RequestException as e:
        return False, str(e)
