## 2026-04-14 - Fix Local Log File Creation Vulnerability
**Vulnerability:** The application was creating a log file in the current working directory using a predictable filename (`ticker-log-YYYY-MM-DD.log`).
**Learning:** This exposes the application to Symlink Attacks (CWE-377 / CWE-59) or directory pollution, especially if run in a shared temporary directory.
**Prevention:** Instead of relying on the implicit current working directory, use secure, user-specific locations like `xdg.StateHome` (via `github.com/adrg/xdg`) to store log files, ensuring the parent directory has strict permissions (e.g., `0700`).
## 2026-04-18 - Enforce Handshake Timeout for Custom WebSockets but Preserve Proxy Config
**Vulnerability:** Replacing `websocket.DefaultDialer` with a custom `&websocket.Dialer{}` to add strict configurations (e.g., `HandshakeTimeout: 10 * time.Second`) drops the inherited default proxy configurations, inadvertently breaking connections for users on proxy networks.
**Learning:** Gorilla WebSocket's `DefaultDialer` automatically initializes `Proxy: http.ProxyFromEnvironment` as well as a 45s HandshakeTimeout. Rebuilding the dialer from scratch removes the proxy setting entirely.
**Prevention:** Always copy the `Proxy` setting from `websocket.DefaultDialer` (e.g., `Proxy: websocket.DefaultDialer.Proxy`) when creating custom dials for stricter security bounds or explicitly handle proxies manually.
