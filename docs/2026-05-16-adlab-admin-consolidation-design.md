# AdLab Admin Consolidation Design

Date: 2026-05-16

## Goal

Consolidate the AdLab admin experience around a single frontend:

- Keep `/Users/redamancy/Dev/AeroBid/adlab-server/admin-frontend` as the only admin UI.
- Remove the incomplete root-level admin frontend.
- Align the surviving admin with the existing backend capabilities under `/admin/*` and `/api/v1/*`.
- Improve the visual consistency and perceived quality of the admin without turning it into a high-risk rewrite.

## Product Direction

This admin should support two equally important workflows:

1. Configuration management
   - Apps
   - Placements
   - Sources
   - DSP configs
   - Materials
   - Mock ads
   - Scenario switching

2. Operational analysis and debugging
   - Dashboard
   - Analytics
   - Bid logs
   - Tracking chains
   - Ad player
   - Change logs
   - Import/export and maintenance tools

The implementation should balance both sides instead of optimizing only for one.

## Constraints

- Reuse the existing Vite + React + Ant Design stack in `adlab-server/admin-frontend`.
- Follow the backend contract as implemented today rather than redesigning backend APIs.
- Prefer medium-risk refactors over a visual-only pass or a full rewrite.
- Keep the default local dev workflow on port `3000`.

## Current Problems

### Structural

- There were two admin frontends with overlapping purpose.
- The root-level admin frontend was incomplete and created ambiguity.
- The retained admin already covers many domains, but the page quality is uneven.

### Contract / Data

- Several backend resources are keyed by business IDs rather than numeric primary keys.
- Some frontend pages already handle backend pagination correctly via `items`, but the contract usage is not fully centralized.
- Types drift from backend fields in places, especially around tracking events, materials, and some log/status shapes.

### UX / Visual

- The shell is decent but not yet fully cohesive.
- Toolbars, card headers, table actions, forms, and empty states are not fully standardized.
- Some pages feel more polished than others.
- Settings and audit flows expose backend power, but not yet with strong guidance and maintenance affordances.

## Recommended Approach

Use a balanced medium refactor with these pillars:

1. Contract-first cleanup
   - Normalize frontend API helpers and shared types to match backend responses and identifiers.
   - Reduce ad hoc response parsing inside pages.

2. Shared UI system refinement
   - Strengthen reusable card, toolbar, section header, stat card, detail drawer, and status patterns.
   - Keep the current dark sidebar + light content layout, but tighten spacing, hierarchy, and consistency.

3. Page consolidation
   - Bring configuration pages and analytics pages onto the same interaction standard.
   - Improve settings and maintenance workflows so backend capabilities are discoverable and safe to use.

4. Progressive enhancement
   - First make pages correct and consistent.
   - Then improve density, clarity, and polish.

## Information Architecture

### Keep

- Overview
- Analytics
- Apps
- Ad Units
- Networks
- DSP Config
- Materials
- Mock Ads
- Bid Logs
- Scenarios
- Ad Player
- Audit Log
- Settings

### Remove

- The deleted root-level admin frontend should remain removed and unrecoverable in the normal workflow.

## UI Direction

### Visual language

- Preserve the dark left navigation and light operational workspace.
- Use the existing orange brand accent as a controlled highlight, not as a dominant fill color.
- Standardize card edges, table framing, filters, and metadata chips.
- Make list pages feel like product tools rather than raw CRUD tables.

### Layout rules

- Each page should use a predictable stack:
  - page shell
  - summary / explanation
  - toolbar / filters
  - primary data surface
  - details / secondary surfaces

- Cards should be used to separate jobs, not decorate everything.
- Empty states should tell the user what to do next.

### Interaction rules

- Destructive actions remain behind confirmation.
- Drawers and modals should use consistent titles, subtitles, and footer actions.
- Common actions should stay in the same place across pages.

## Functional Upgrades

### Shared

- Centralize backend response handling.
- Tighten typed resource IDs.
- Improve reusable components and shared styling tokens.

### Configuration management

- Apps: better card/list readability, clearer mock fallback control, stronger placement linkage.
- Placements: cleaner network binding workflow, better source discovery, stronger details drawer.
- Sources: expose advanced network fields clearly and progressively.
- DSP configs: maintain simulator power while making bid behavior easier to scan and edit.
- Materials: align media editing with backend update support.
- Mock ads: keep preview-first workflow, improve density and filtering.
- Scenarios: make current state, impact, and refresh behavior more explicit.

### Analysis and debugging

- Dashboard: stronger summary, cleaner quick-test area, more operational value.
- Analytics: improved readability and filtering continuity.
- Bid logs: better filter affordances, clearer drawers, stronger tracking and detail presentation.
- Tracking timeline: normalize event mapping and improve readability.
- Ad player: keep as a debugging workstation, but visually integrate it with the rest of the admin.
- Audit log: surface diff information more clearly and add filter readiness.

### Maintenance tools

- Settings should expand beyond import/export display into a real admin operations panel.
- Add safe access to:
  - config export
  - config import
  - seed data
  - log cleanup
  - environment / storage information

## Implementation Sequence

1. Document and lock scope
2. Normalize shared API client, types, and reusable UI components
3. Refine shell and global styling
4. Upgrade high-frequency config pages
5. Upgrade high-frequency analytics and debugging pages
6. Expand settings and maintenance tools
7. Run build and browser verification

## Risks

- Some pages currently depend on slightly different assumptions about backend field names.
- Shared type cleanup can cause ripple edits across many pages.
- Settings and maintenance tools need careful UX because they expose powerful backend operations.

## Non-Goals

- No backend API redesign as part of this pass.
- No charting library migration unless existing simple solutions block clarity.
- No theme system expansion into multiple themes or dark-mode content area redesign.

## Success Criteria

- Only one admin frontend remains in the repo and workflow.
- The retained admin builds and runs on port `3000`.
- Core config pages and analysis pages are both aligned to backend behavior.
- The UI feels like one product rather than a collection of partially matched pages.
- Settings exposes the backend maintenance capabilities in a more complete, safer way.
