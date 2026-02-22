"""TinyStock - Stock tracking web app with JWT auth."""

import streamlit as st

from api_client import login, register

st.set_page_config(
    page_title="TinyStock",
    page_icon="📈",
    layout="wide",
    initial_sidebar_state="expanded",
)

# Session state for auth
if "token" not in st.session_state:
    st.session_state.token = None
if "user" not in st.session_state:
    st.session_state.user = None


def render_login():
    """Render login/register form."""
    tab1, tab2 = st.tabs(["Login", "Register"])
    with tab1:
        with st.form("login_form"):
            email = st.text_input("Email", placeholder="demo@tinystock.app")
            password = st.text_input("Password", type="password", placeholder="demo123")
            col1, col2 = st.columns(2)
            with col1:
                submitted = st.form_submit_button("Login")
            with col2:
                demo = st.form_submit_button("Demo Login")
            if demo:
                email, password = "demo@tinystock.app", "demo123"
                submitted = True
            if submitted and email and password:
                with st.spinner("Login..."):
                    ok, token, user = login(email, password)
                    if ok:
                        st.session_state.token = token
                        st.session_state.user = user
                        st.success("Logged in!")
                        st.rerun()
                    else:
                        st.error("Invalid email or password")
    with tab2:
        with st.form("register_form"):
            reg_email = st.text_input("Email", key="reg_email")
            reg_password = st.text_input("Password", type="password", key="reg_password")
            if st.form_submit_button("Register"):
                if reg_email and reg_password:
                    if len(reg_password) < 6:
                        st.error("Password must be at least 6 characters long")
                        return
                    with st.spinner("Registering..."):
                        ok, msg = register(reg_email, reg_password)
                        if ok:
                            st.success(msg)
                        else:
                            st.error(msg)


# Main app
if st.session_state.token is None:
    st.title("📈 TinyStock")
    st.caption("Stock quotes, watchlist, and portfolio tracking")
    st.divider()
    render_login()
else:
    st.title("📈 TinyStock")
    user = st.session_state.user or {}
    st.caption(f"Logged in as {user.get('email', '')}")
    if st.sidebar.button("Logout"):
        st.session_state.token = None
        st.session_state.user = None
        st.rerun()
    st.sidebar.divider()
    page = st.sidebar.radio(
        "Navigate",
        ["Dashboard", "Quote Lookup", "Watchlist", "Portfolio"],
        label_visibility="collapsed",
    )
    if page == "Dashboard":
        from views import dashboard
        dashboard.render(st.session_state.token)
    elif page == "Quote Lookup":
        from views import quote
        quote.render(st.session_state.token)
    elif page == "Watchlist":
        from views import watchlist
        watchlist.render(st.session_state.token)
    elif page == "Portfolio":
        from views import portfolio
        portfolio.render(st.session_state.token)
