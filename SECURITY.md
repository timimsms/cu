# Security Policy

`cu` is an unofficial, community-maintained project and is not affiliated with or endorsed by ClickUp.

## Token Storage

`cu` stores your ClickUp personal API token in the operating system's native credential store via [zalando/go-keyring](https://github.com/zalando/go-keyring), under the service name `cu-cli`:

- **macOS**: Keychain
- **Windows**: Credential Manager
- **Linux**: Secret Service (e.g. GNOME Keyring, KWallet)

There is no plaintext fallback — tokens are never written to disk in plaintext by `cu`. On headless Linux systems, a Secret Service implementation must be available or authentication will fail.

## Revoking Access

If a token is compromised or no longer needed:

1. Revoke it in ClickUp: **Settings → Apps**, then delete the token.
2. Remove the stored copy locally: `cu auth logout`

## Reporting a Vulnerability

Please report security vulnerabilities privately by emailing tim@mims.ms rather than opening a public issue. Include a description of the issue, steps to reproduce, and any relevant details. You should receive a response within a few days.
