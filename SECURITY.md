# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | Yes |
| Older releases | No |

We recommend always using the latest version. Run `magecli update` to self-update.

## Reporting a Vulnerability

If you discover a security vulnerability in magecli, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email security concerns to the maintainers via the contact information on the [Atlantic BT GitHub organization](https://github.com/atlanticbt).

### What to include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### What to expect

- Acknowledgment within 48 hours
- Status update within 7 days
- We will coordinate disclosure timing with you

## Scope

This policy applies to the magecli CLI tool itself. Issues with the Magento 2 REST API should be reported to Adobe/Magento directly.

## Security Design

- **No credentials in config files** — Bearer tokens are stored in the OS keyring, never written to disk
- **Read-only by default** — Contexts block write operations unless explicitly enabled with `allow_writes: true`
- **No shell execution** — magecli does not execute shell commands or evaluate user input as code
- **TLS by default** — All API communication uses HTTPS
