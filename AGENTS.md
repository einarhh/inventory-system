# AGENTS.md

Read this file at the start of every session. It defines the architecture, the
conventions, and — most importantly — the **scope boundary**. Do not build
anything marked Phase 2+ unless explicitly told to in the current task.

## Commit conventions

Do **not** add Claude/AI attribution to commits or PRs. No
`Co-Authored-By: Claude` trailers and no "Generated with Claude Code" footers.

## What this is

A personal/collective inventory system. A backend service (Go) exposes a REST
API over a PostgreSQL database. A web UI consumes that API. A mobile app may come
later. The defining feature is a **nested location tree**: items live somewhere,
and "somewhere" can be Area → Building → Floor → Room → Rack → Cabinet → Box →
Compartment → Vehicle, nested arbitrarily deep.

## Architecture

- **Language:** Go (1.22+).
- **HTTP:** `net/http` with the 1.22 routing enhancements (`mux.HandleFunc("GET /items/{id}", ...)`). No heavy framework. `chi` is acceptable only if middleware needs grow.
- **DB:** PostgreSQL, accessed via `pgx` (v5). SQL is written by hand and compiled to type-safe Go with **sqlc**. We do NOT use a general-purpose ORM.
- **Migrations:** `goose`, plain `.sql` files in `db/migrations/`.
- **Config:** environment variables only. No config files with secrets.
- **Tests:** stdlib `testing`. Integration tests use `testcontainers-go` to spin up a real Postgres. Prefer testing against real SQL over mocks.
- **Frontend:** React + Vite + TypeScript, built as an installable PWA (web app manifest + service worker). Talks to the Go API over REST only.

## Project layout

```
cmd/server/        main.go — wiring, config, server start
internal/
  http/            handlers, routing, middleware, request/response types
  store/           sqlc-generated code + hand-written queries (queries.sql)
  domain/          core types (Location, Item, Customer, File) and validation
db/migrations/     goose migrations
docs/              SPEC.md and other design docs
```

Keep `internal/` truly internal. Handlers depend on the store; the store depends
on the DB. Domain types have no DB or HTTP imports.

## Conventions

- Every endpoint that mutates data validates input in `domain` before touching the store.
- Errors returned to clients are JSON `{ "error": "message" }` with a sensible status code. Never leak SQL errors verbatim.
- Every new endpoint ships with at least one integration test in the same PR/commit.
- Migrations are append-only. Never edit a migration that has been committed; write a new one.
- IDs are UUIDs (generated DB-side with `gen_random_uuid()`).
- Money is stored in minor units (integer cents) — never floats.
- Timestamps are `timestamptz`, UTC.

## How we work

Build in small, reviewable increments. The order is:
1. Schema / migration.
2. sqlc queries + generated code.
3. One endpoint + its test.
4. Repeat per endpoint.
5. Then the web UI, one view at a time.

Do not scaffold the entire API in one pass. Stop after each increment so it can
be reviewed.

## Scope boundary

### In scope (MVP)
- Nested location tree (single parent per location).
- Items belonging to exactly one location.
- Item fields: name, quantity, fill level (0–100), capacity, condition, date acquired, purchase price, photo, receipt.
- CRUD: customer, location, item, file/image.
- Statistics endpoint (item count, total value) with filters.
- Plain-text search over item name/notes (SQL `ILIKE` or `tsvector`; no Elasticsearch).
- Web UI: a mobile-first, installable PWA (browse the location tree, view/edit items, attach photos via the camera/file APIs, search) that also works as a desktop web app.

### Phase 2+ (DO NOT build unless explicitly asked)
Sharing & lending, marketplace/bids, ratings, messaging, BLE/beacons/IoT,
addressable LEDs, AR, gamification/points, GraphQL, SSO/MFA, monetization tiers,
label printing, fleet management, sustainability tracking, voice assistants,
Excel import/export, reports, GPS trackers, computer vision.

These are real future goals — the data model leaves room for them (see "Extensibility")
— but they are not part of the first working version.

## Extensibility (design now, build later)

The note author wants a clean single-parent tree now, with the ability to add
other relationships later. Therefore:
- Location parentage is a single nullable `parent_id` on `locations`. **Do not**
  model the tree with a closure table or array yet.
- When arbitrary relationships are needed later (item↔item "charger belongs to
  laptop", item-acts-as-location, multi-parent), add a separate
  `relationships` join table rather than altering the tree. Leave the core tables
  unpolluted so this is additive, not a rewrite.
- Same physical item type in multiple places is handled by separate item rows for
  now; a `sku`/`product` concept can be layered on later via a nullable column.
- The frontend stays a plain REST client so a native app can be added later by
  either wrapping the PWA (e.g. Capacitor) or building a separate native client
  against the same API. Native is deferred until features that browsers can't do
  are scheduled — specifically BLE beacons and NFC (Web Bluetooth / Web NFC are
  not supported on iOS Safari), which already live in the Phase 2+ IoT list.
