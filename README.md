![lynway logo](docs/logo.png)

# Lynway WMS backend service (v1.0-alpha)

[Русская версия](README_ru.md)

**Lightweight warehouse management system (WMS). For individual storage or small businesses.**

This repository contains the **Community Edition (CE)** — the open-source backend of Lynway WMS, distributed under the **AGPLv3** license (see [LICENSE](LICENSE)).

The paid editions — **Business** and **Cloud** — are developed in a private repository and are not published on GitHub.

* **Swagger UI:** http://localhost:8888/api/v1/docs/index.html

---

## 🏗 Tech Stack & Architecture

* **Backend:** Go (`database/sql`, `lib/pq`, `golang-migrate` with `embed.FS`)
* **Database:** PostgreSQL 16 (`pg_trgm`, `JSONB`)
* **Clients (in progress):**
    * **Android App:** QR/Barcode scanning, container inspection, stock lookup, and internal moves.
    * **Vue 3 SPA:** Web UI for receiving, smart search, and topology management.
    * **CUPS Print Agent:** Local daemon syncing CUPS printers and processing label print jobs.

---

## 🚀 Deploy for Development

1. Clone the repository and start the database in Docker:
   ```bash
   sudo docker compose up -d
   ```
2. Run the service (the database will be initialized automatically on first start):
   ```bash
   make run
   ```
   Default administrator credentials: `admin` / `admin`.


## 📦 Product Editions

| Edition | Code | License | Location |
| :--- | :--- | :--- | :--- |
| Community Edition | open source | AGPLv3 | this repository |
| Business | closed source | commercial | private repository |
| Cloud | closed source | SaaS / commercial | private repository |

### How we develop without edition conflicts

1. **Shared core.** The CE repository is the source of truth for base logic. Business/Cloud are built on top of it as an overlay, not as a fork with separate branches.
2. **Private repo + synchronization.** The enterprise repository consumes CE code (submodule or allowlist-based sync script). Paid modules live only there and never end up on GitHub.
3. **Build tags.** Commercial code is isolated with `//go:build enterprise`; such files are simply ignored by the compiler in CE builds. Extension points are interfaces and a registry: in CE the registry is empty.
4. **CI guarantee of CE purity.** CI in this repository builds and tests the binary **without** the `enterprise` tag — this prevents proprietary code from leaking into the open build.
5. **CLA.** CE contributors sign a Contributor License Agreement so we retain dual-licensing rights (ability to include their code in Business/Cloud).

---

## 🗺 Roadmap Status (v1.0-alpha)

* [x] Core database schema
* [x] Database auto-creation, embedded migrations (`golang-migrate`), and initial seeding
* [ ] Append-only ledger engine (`storage`) and zone/cell restriction checks
* [ ] Nested container math (`containers`: nest, unnest, deep inspection)
* [ ] Item resolution service ("Find, Clone or Create" + `pg_trgm`)
* [ ] REST API for Android app, Vue 3 SPA, and CUPS Print Agent
