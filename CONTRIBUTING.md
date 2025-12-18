
<sub>🇩🇪 [German translation →](CONTRIBUTING.de.md)</sub>


# Contributing to the GSF Suite

First of all: **Thank you** for taking the time to contribute to the **GSF Suite**.
Open source lives from people who share their knowledge, time, and experience.

This document outlines a few guidelines intended to keep the project **clear, stable, and maintainable in the long term**.

---

## 1. Reporting Bugs

Before opening a new issue, please check:

* whether the problem has already been reported
* whether you are using the **latest available version** of the relevant module

When reporting a bug, the following information is very helpful:

* version(s) used
* steps to reproduce the issue
* expected behavior vs. actual behavior
* optional: a **minimal code example**

> **Note on security issues:**
> If you believe the issue may be **security-related**, please **do not open a public issue**,
> but follow the instructions in `SECURITY.md`.

---

## 2. Pull Requests (Contributing Code)

We welcome pull requests — whether they fix bugs, improve existing code, or add new features.

To help us review and merge your PR efficiently, please follow these guidelines:

1. **Fork & Branch**
   Fork the repository and work in a dedicated feature branch:

   ```bash
   git checkout -b feature/my-feature
   ```

2. **Coding Style**
   Please follow the existing coding style and design principles of the respective module.

3. **Tests**

   * New functionality should be accompanied by appropriate tests
   * Bug fixes should include a regression test when reasonable

4. **License Header**
   New source files must include the correct SPDX license header:

   ```go
   // SPDX-License-Identifier: Apache-2.0
   ```

5. **Security-related Changes**
   If your pull request addresses a potential security vulnerability,
   please refer to `SECURITY.md` and consider submitting the PR as a **draft** first.

---

## 3. Legal & Licensing

By submitting a pull request, you confirm that:

1. you are the original author of the contributed code **or** have the right to submit it
2. your contribution may be published under the **Apache License 2.0**

This project follows the **"Inbound = Outbound"** principle:

* No separate Contributor License Agreement (CLA) is required
* All contributions are licensed under the same terms as the project itself

---

## 4. Philosophy

The GSF Suite deliberately follows a **Tiny / Simple philosophy**:

* minimal dependencies
* explicit APIs over magic
* small, clearly scoped modules
* predictability over feature richness

Contributions should respect and align with these principles.

---

Thank you for your contribution!
**We** appreciate constructive discussions, clean contributions, and shared evolution.

—
Georg Hagn

