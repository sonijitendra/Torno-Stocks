"""Dashboard - overview of watchlist and portfolio."""

import streamlit as st

from api_client import get_watchlist, get_portfolio


def render(token: str):
    st.header("Dashboard")

    # Watchlist summary
    st.subheader("Watchlist")
    watchlist_items, quotes = get_watchlist(token)

    if watchlist_items:
        # Build quote map
        quote_map = {q.get("symbol"): q for q in quotes if isinstance(q, dict)}

        n = len(watchlist_items)
        cols = st.columns(min(n, 4))
        for i, item in enumerate(watchlist_items):
            symbol = item.get("symbol", "")
            q = quote_map.get(symbol, {})
            price = q.get("price", 0)
            change_pct = q.get("changePercent", 0)
            with cols[i % 4]:
                st.metric(
                    symbol,
                    f"${price:,.2f}" if price else "N/A",
                    delta=f"{change_pct:+.2f}%" if change_pct else None,
                )
    else:
        st.info("No stocks in watchlist. Add some from the Quote Lookup or Watchlist page.")

    st.divider()

    # Portfolio summary
    st.subheader("Portfolio Summary")
    portfolio = get_portfolio(token)
    holdings = portfolio.get("holdings", [])
    total_value = portfolio.get("totalValue", 0)
    total_cost = portfolio.get("totalCost", 0)
    total_pnl = portfolio.get("totalPnL", 0)

    if holdings:
        pnl_pct = portfolio.get("returnPct", 0) or (((total_value - total_cost) / total_cost * 100) if total_cost > 0 else 0)
        col1, col2, col3, col4 = st.columns(4)
        with col1:
            st.metric("Total Value", f"${total_value:,.2f}")
        with col2:
            st.metric("Total Cost", f"${total_cost:,.2f}")
        with col3:
            st.metric("P&L", f"${total_pnl:,.2f}", delta=f"{pnl_pct:+.2f}%")
        with col4:
            st.metric("Holdings", len(holdings))

        st.subheader("Holdings")
        for h in holdings:
            col1, col2, col3, col4 = st.columns([1, 2, 2, 2])
            with col1:
                st.write(f"**{h.get('symbol', '')}**")
            with col2:
                st.write(f"{h.get('quantity', 0):.2f} @ ${h.get('buyPrice', 0):,.2f}")
            with col3:
                st.write(f"Value: ${h.get('marketValue', 0):,.2f}")
            with col4:
                pnl = h.get("pnl", 0)
                pnl_pct = h.get("pnlPercent", 0)
                st.write(f"P&L: ${pnl:,.2f} ({pnl_pct:+.2f}%)")
    else:
        st.info("No holdings. Add some from the Portfolio page.")
