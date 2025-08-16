# CARES — Phase 01 TODO

This file lists the components and tasks for Phase 01 (Standalone, Self-Aware Node). Tasks are grouped and ordered to help implementation in small, testable steps.

## Core modules

- [ ] Verify existing `metrics` module
  - [ ] Confirm exported API and types
  - [ ] Add simple unit tests for `GetCPUUsage` and `GetMemoryUsage`
  - [ ] Ensure error values are returned (not printed)
- [ ] TUI (Bubble Tea)
  - [ ] Add Bubble Tea dependency to project
  - [ ] App bootstrap (program entrypoint)
  - [ ] Define `Model` (holds CPU/memory and UI state)
  - [ ] Implement `Update` loop handling ticks and input
  - [ ] Implement `View` renderers (bars, numeric displays)
  - [ ] Key handlers (quit, toggle debug, refresh)
  - [ ] Integrate config/flags for refresh interval

## Concurrency & integration

- [ ] Metrics collector goroutine (periodic sampling)
- [ ] Channel(s) to deliver metric samples to the TUI
- [ ] Implement tick / timer coordination for redraws

## Configuration & CLI

- [ ] Add CLI flags: `--interval`, `--no-tui`, `--debug`
- [ ] Provide reasonable default values
- [ ] Support environment variable overrides

## Reliability & UX

- [ ] Graceful shutdown (handle SIGINT/SIGTERM)
- [ ] Display errors in TUI (user-friendly messages)
- [ ] Fallback UI state when metrics unavailable (show `N/A`)

## Observability & tooling

- [ ] Add basic logging (configurable verbosity)
- [ ] Provide a `--dev` mode that prints metrics to stdout (non-TUI)

## Project hygiene

- [ ] Update README with Phase 01 run instructions
- [ ] Document dependencies (Bubble Tea, gopsutil)
- [ ] Add basic unit tests for metrics functions
- [ ] Create example commands for building and running

## Prioritization (Suggested order)

1. Verify `metrics` API and return values
2. Make metrics collector goroutine + channel
3. Implement minimal Bubble Tea app skeleton (bootstrap + model)
4. Wire metrics -> TUI (display numeric values)
5. Add graceful shutdown and CLI flags
6. Improve UI visuals (bars, layout, keybindings)
7. Add tests, docs, and polishing

---

If you want, I can split these items into individual issues or create a tracked checklist file per component next.
