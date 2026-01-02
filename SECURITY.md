
<sub>🇩🇪 [German translation →](SECURITY.de.md)</sub>


# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability, please **do not open a public issue**.

Instead, report it responsibly:

* Contact the maintainer via GitHub (private message)
* Alternatively: Send an email to the address in the README
* Optional: open a draft pull request describing the issue and proposed fix

Please include:

* A clear description of the vulnerability
* Steps to reproduce (if applicable)
* Affected module(s)
* Any suggested mitigation or fix

For general bug reports, feature suggestions, or questions, please use the regular GitHub issues according to the instructions in `CONTRIBUTING.md`.

---

## Scope

GSF focuses on:

* In-process libraries (logging, scheduling, file rotation)
* Minimal dependencies (standard library preferred)

Out of scope:

* Network security (TLS, authentication, authorization)
* OS-level hardening
* Application-level security policies

---

## Disclosure Policy

Reported vulnerabilities will be:

1. Reviewed as soon as reasonably possible
2. Fixed in a private branch if necessary
3. Released publicly once a fix is available

There is no formal CVE process at this stage.

---

## Final Note

GSF aims to be **simple, predictable, and transparent**.

If something looks unsafe, ambiguous, or surprising: **please report it**

Security arises from shared awareness.

---

