"""Watchlist page."""

import streamlit as st

from api_client import get_watchlist, add_to_watchlist, remove_from_watchlist


def render(token: str):
    st.header("Watchlist")

    # Add symbol
    with st.expander("Add to Watchlist"):
        new_symbol = st.text_input("Symbol", placeholder="e.g. AAPL", key="watchlist_add")
        if st.button("Add"):
            if new_symbol:
                success, msg = add_to_watchlist(token, new_symbol.strip().upper())
                if success:
                    st.success(msg)
                    st.rerun()
                else:
                    st.error(msg)
            else:
                st.warning("Enter a symbol")

    # List watchlist
    watchlist_items, quotes = get_watchlist(token)
    quote_map = {q.get("symbol"): q for q in quotes if isinstance(q, dict)}

    if watchlist_items:
        st.subheader("Your Watchlist")
        for i, w in enumerate(watchlist_items):
            symbol = w.get("symbol", "")
            q = quote_map.get(symbol, {})
            price = q.get("price", 0)
            change = q.get("change", 0)
            change_pct = q.get("changePercent", 0)

            col1, col2, col3, col4, col5 = st.columns([2, 2, 2, 2, 1])
            with col1:
                st.write(f"**{symbol}**")
            with col2:
                st.write(f"${price:,.2f}" if price else "N/A")
            with col3:
                st.write(f"${change:,.2f}" if change else "-")
            with col4:
                st.write(f"{change_pct:+.2f}%" if change_pct else "-")
            with col5:
                if st.button("Remove", key=f"rm_{i}_{symbol}"):
                    remove_from_watchlist(token, symbol)
                    st.rerun()
    else:
        st.info("Your watchlist is empty. Add stocks from the Quote Lookup page.")
