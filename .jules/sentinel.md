## 2026-04-14 - Fix Local Log File Creation Vulnerability
**Vulnerability:** The application was creating a log file in the current working directory using a predictable filename (`ticker-log-YYYY-MM-DD.log`).
**Learning:** This exposes the application to Symlink Attacks (CWE-377 / CWE-59) or directory pollution, especially if run in a shared temporary directory.
**Prevention:** Instead of relying on the implicit current working directory, use secure, user-specific locations like `xdg.StateHome` (via `github.com/adrg/xdg`) to store log files, ensuring the parent directory has strict permissions (e.g., `0700`).
