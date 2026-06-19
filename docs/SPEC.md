# SPEC.md — MVP

This is the requirements document for the first working version. It is derived
from several years of notes. The vision is much larger (see the Phase 2+ list in
AGENTS.md); this document is deliberately small so it can actually ship.

## Goal

Let one user catalog their belongings by where they are, find things quickly, and
see what they own and what it's worth. Everything else is later.

## Core concepts

**Customer** — the account that owns data. (Single-tenant feel for now, but the
column exists so multi-user is not a rewrite.)

**Location** — a place that can contain items and other locations. Has a `type`
(area, building, floor, room, rack, cabinet, box, compartment, vehicle, other)
and an optional single parent location. Forms a tree.

**Item** — a thing that lives in exactly one location.

**File** — an image or document (e.g. receipt) attached to an item.

## Functional requirements

### Locations
- Create, read, update, delete a location.
- A location may have a parent (same customer). Deleting a location with children
  or items is rejected unless `?cascade=true` is passed.
- Fetch the tree (or a subtree) for a customer.

### Items
- Create, read, update, delete an item.
- An item must reference a valid location owned by the same customer.
- Fields: name (required), quantity (default 1), fill_level (0–100, nullable),
  capacity (nullable), condition (new/good/fair/poor, nullable), date_acquired
  (nullable), purchase_price_cents (nullable), notes (nullable).
- Move an item to a different location.

### Files
- Upload an image/document, attach to an item.
- For MVP, store files on local disk (or a configured dir); persist metadata +
  path in DB. Object storage is a later swap behind the same interface.

### Search
- Plain-text search over item name and notes for a customer. `ILIKE` is fine for
  MVP; a `tsvector` column is acceptable if trivial.

### Statistics
- For a customer (optionally filtered by subtree/location): total item count,
  total quantity, and total value (sum of purchase_price_cents × quantity).

### Web UI (after the API is solid)
- Browse the location tree.
- View a location's items.
- Add/edit an item, attach a photo/receipt.
- A simple search box.

## Non-functional
- API returns JSON. Errors are `{ "error": "..." }`.
- All list endpoints support basic pagination (`?limit=&offset=`).
- Input validated server-side; never trust the client.

## Explicitly out of scope for MVP
See AGENTS.md "Phase 2+". If a task seems to require any of those, stop and ask.
