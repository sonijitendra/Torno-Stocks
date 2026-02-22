"""Quote lookup page."""

import streamlit as st

from api_client import get_quote, get_history, search_symbols, add_to_watchlist


def render(token: str):
    st.header("Stock Quote Lookup")

    # Search or direct symbol input
    col1, col2 = st.columns([2, 1])
    with col1:
        query = st.text_input("Search or enter symbol", placeholder="e.g. AAPL, MSFT, GOOGL")
    with col2:
        if st.button("Look Up", type="primary"):
            st.session_state["lookup_symbol"] = query.upper() if query else None

    # Show search results if query looks like a search (multiple words or partial)
    is_search = query and (
        (len(query) > 2 and " " in query) or (len(query) > 1 and not query.isupper())
    )
    if is_search:
        results = search_symbols(query, 5)
        if results:
            st.subheader("Search Results")
            for r in results:
                if st.button(f"{r.get('symbol', '')} - {r.get('name', '')}", key=f"search_{r.get('symbol')}"):
                    st.session_state["lookup_symbol"] = r.get("symbol", "").upper()
            st.divider()

    # Display quote
    symbol = st.session_state.get("lookup_symbol") or (query.upper() if query and query.isupper() and len(query) <= 5 else None)

    if symbol:
        quote = get_quote(symbol)
        if quote:
            st.subheader(f"{quote.get('symbol', symbol)} - {quote.get('name', '')}")

            col1, col2, col3, col4 = st.columns(4)
            with col1:
                st.metric("Price", f"${quote.get('price', 0):,.2f}")
            with col2:
                change = quote.get("change", 0)
                change_pct = quote.get("changePercent", 0)
                st.metric("Change", f"${change:,.2f} ({change_pct:+.2f}%)", delta=f"{change_pct:.2f}%")
            with col3:
                st.metric("Volume", f"{quote.get('volume', 0):,}")
            with col4:
                st.metric("Day Range", f"${quote.get('low', 0):,.2f} - ${quote.get('high', 0):,.2f}")

            # Add to watchlist
            if st.button("Add to Watchlist"):
                success, msg = add_to_watchlist(token, symbol)
                if success:
                    st.success(msg)
                else:
                    st.warning(msg)

            # Historical chart
            st.subheader("Price History (30 days)")
            history = get_history(symbol, "1mo", "1d")
            if history:
                import pandas as pd
                import plotly.express as px

                df = pd.DataFrame(history)
                df["date"] = pd.to_datetime(df["date"])
                fig = px.line(df, x="date", y="close", title=f"{symbol} - Last 30 Days")
                fig.update_layout(height=400, margin=dict(l=0, r=0))
                st.plotly_chart(fig, use_container_width=True)
            else:
                st.info("Historical data unavailable")
        else:
            st.error(f"Could not fetch quote for {symbol}. Is the backend running?")
    else:
        st.info("Enter a stock symbol (e.g. AAPL) and click Look Up")
