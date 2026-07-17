# Authentication & Security

`cu` authenticates with ClickUp using a personal API token.

## Creating a ClickUp Personal API Token

1. Log in to ClickUp and open **Settings → Apps** (https://app.clickup.com/settings/apps).
2. Under **API Token**, click **Generate** (or **Regenerate**).
3. Copy the token — it starts with `pk_`.

Then authenticate:

```bash
cu auth login
```

Paste the token when prompted, or pass it directly:

```bash
cu auth login --token pk_xxxxxxxx
```

Check your status at any time with `cu auth status`.

## Where Your Token Is Stored

`cu` stores the token in your operating system's native credential store via the [zalando/go-keyring](https://github.com/zalando/go-keyring) library, under the service name `cu-cli`:

| OS | Credential store |
| --- | --- |
| macOS | Keychain |
| Windows | Credential Manager |
| Linux | Secret Service (GNOME Keyring, KWallet, etc.) |

Tokens are never written to disk in plaintext by `cu` — there is no plaintext fallback.

## Headless Linux Caveat

On headless Linux (servers, containers, CI), `cu` requires a running Secret Service implementation to store and read the token. There is currently no environment-variable fallback, so authentication will fail without one. A common workaround is to run a keyring daemon such as `gnome-keyring-daemon` with a D-Bus session. An environment-variable token fallback is planned — see [issue #26](https://github.com/timimsms/cu/issues/26).

## Revoking Access

If a token is compromised or no longer needed:

1. Revoke it in ClickUp under **Settings → Apps** so it can no longer be used anywhere.
2. Remove the locally stored copy:

```bash
cu auth logout
```

## Reporting Security Issues

See the project's [security policy](https://github.com/timimsms/cu/blob/main/SECURITY.md) for how to report vulnerabilities.
