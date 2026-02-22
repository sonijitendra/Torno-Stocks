"""Portfolio page."""

import streamlit as st

from api_client import get_portfolio, add_holding, remove_holding


def render(token: str):
    st.header("Portfolio")

    # Add holding
    with st.expander("Add Holding"):
        col1, col2, col3 = st.columns(3)
        with col1:
            symbol = st.text_input("Symbol", placeholder="e.g. AAPL", key="portfolio_symbol")
        with col2:
            quantity = st.number_input("Quantity", min_value=0.01, value=1.0, step=0.01, key="portfolio_qty")
        with col3:
            buy_price = st.number_input("Buy Price ($)", min_value=0.01, value=100.0, step=0.01, key="portfolio_price")

        if st.button("Add Holding"):
            if symbol:
                success, msg = add_holding(token, symbol.strip().upper(), quantity, buy_price)
                if success:
                    st.success(msg)
                    st.rerun()
                else:
                    st.error(msg)
            else:
                st.warning("Enter a symbol")

    # Portfolio summary
    portfolio = get_portfolio(token)
    holdings = portfolio.get("holdings", [])
    total_value = portfolio.get("totalValue", 0)
    total_cost = portfolio.get("totalCost", 0)
    total_pnl = portfolio.get("totalPnL", 0)

    if holdings:
        pnl_pct = portfolio.get("returnPct", 0) or (((total_value - total_cost) / total_cost * 100) if total_cost > 0 else 0)

        st.subheader("Summary")
        col1, col2, col3 = st.columns(3)
        with col1:
            st.metric("Total Value", f"${total_value:,.2f}")
        with col2:
            st.metric("Total Cost", f"${total_cost:,.2f}")
        with col3:
            st.metric("Total P&L", f"${total_pnl:,.2f}", delta=f"{pnl_pct:+.2f}%")

        st.subheader("Holdings")
        for h in holdings:
            hid = h.get("id")
            symbol = h.get("symbol", "")
            quantity = h.get("quantity", 0)
            buy_price = h.get("buyPrice", 0)
            current_price = h.get("currentPrice", 0)
            market_value = h.get("marketValue", 0)
            cost_basis = h.get("costBasis", 0)
            pnl = h.get("pnl", 0)
            pnl_pct = h.get("pnlPercent", 0)

            with st.container():
                col1, col2, col3, col4, col5, col6 = st.columns([2, 2, 2, 2, 2, 1])
                with col1:
                    st.write(f"**{symbol}**")
                with col2:
                    st.write(f"{quantity:.2f} @ ${buy_price:,.2f}")
                with col3:
                    st.write(f"${current_price:,.2f}")
                with col4:
                    st.write(f"${market_value:,.2f}")
                with col5:
                    st.write(f"${pnl:,.2f} ({pnl_pct:+.2f}%)")
                with col6:
                    if st.button("Remove", key=f"rm_holding_{hid}"):
                        remove_holding(token, hid)
                        st.rerun()
                st.divider()
    else:
        st.info("No holdings. Add your first holding above.")
