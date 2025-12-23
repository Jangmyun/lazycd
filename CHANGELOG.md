# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **TUI Skeleton**: Basic layout with Browser, Shelf, Status bar using `gocui` (Sprint 1).
- **File Browser**:
  - Vim-style navigation (`j`, `k`, `h`, `l`).
  - Directory loading and sorting (directories first).
  - Multi-selection support (Space key).
  - Target selection (`t`).
- **Shelf System**:
  - Add items to shelf (`a`).
  - View shelf items (Tab).
  - Remove items from shelf (`r`).
  - Stale item detection.
- **Persistence**: Save implementation state (last directory, shelf items) to `~/.config/lazycd/state.json`.
- **Navigation policy**: Always start in current working directory (`pwd`).
