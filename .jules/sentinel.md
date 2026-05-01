## 2026-04-14 - Fix Local Log File Creation Vulnerability
**Vulnerability:** The application was creating a log file in the current working directory using a predictable filename (`ticker-log-YYYY-MM-DD.log`).
**Learning:** This exposes the application to Symlink Attacks (CWE-377 / CWE-59) or directory pollution, especially if run in a shared temporary directory.
**Prevention:** Instead of relying on the implicit current working directory, use secure, user-specific locations like `xdg.StateHome` (via `github.com/adrg/xdg`) to store log files, ensuring the parent directory has strict permissions (e.g., `0700`).
## 2026-04-18 - Enforce Handshake Timeout for Custom WebSockets but Preserve Proxy Config
**Vulnerability:** Replacing `websocket.DefaultDialer` with a custom `&websocket.Dialer{}` to add strict configurations (e.g., `HandshakeTimeout: 10 * time.Second`) drops the inherited default proxy configurations, inadvertently breaking connections for users on proxy networks.
**Learning:** Gorilla WebSocket's `DefaultDialer` automatically initializes `Proxy: http.ProxyFromEnvironment` as well as a 45s HandshakeTimeout. Rebuilding the dialer from scratch removes the proxy setting entirely.
**Prevention:** Always copy the `Proxy` setting from `websocket.DefaultDialer` (e.g., `Proxy: websocket.DefaultDialer.Proxy`) when creating custom dials for stricter security bounds or explicitly handle proxies manually.
## 2026-04-25 - Avoid Hardcoded or Missing HTTP Header Values (User-Agent)
**Vulnerability:** In `internal/cli/symbol/symbol.go`, the application was sending HTTP requests without a standard `User-Agent` header, causing it to be exposed to inadvertent blocks by APIs, rate-limiters, or bot protections that check for valid HTTP headers.
**Learning:** This could lead to a localized Denial of Service (DoS) for users whose requests fail consistently because the APIs reject standard Go default behavior. Using variables and constants for standard user-agent strings promotes consistency and maintains uptime.
**Prevention:** Avoid hardcoding HTTP header values inline or missing them entirely. Define explicit constants (e.g., `defaultUserAgent`) and use `http.NewRequest` combined with explicit header assignment (like `req.Header.Set("User-Agent", defaultUserAgent)`) instead of `http.Get()`.
## 2026-04-30 - Fix Missing Input Validation on CSV Parsing
**Vulnerability:** In `internal/cli/symbol/symbol.go`, `parseTickerSymbolToSourceSymbol` parsed a CSV response directly from a URL without validating that the row contained the expected number of columns.
**Learning:** If the CSV is malformed or maliciously crafted (e.g., missing columns), the application will panic when trying to access `row[2]`, resulting in a Denial of Service (DoS) and application crash.
**Prevention:** Always validate the length of slices returned from untrusted or external data parsers (like CSV or JSON arrays) before accessing specific indices by index number.
