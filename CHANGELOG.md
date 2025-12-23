# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Core Operations**:
  - `Copy`: File and recursive directory support with attribute preservation.
  - `Move`: Rename with fallback to Copy+Delete.
  - `Delete`: Move to trash (`~/.config/lazycd/trash`).
- **Conflict Handling**:
  - Detection of existing destinations.
  - Policies: Skip (default), Rename (`name (N).ext`), Overwrite (files only).
- **Job & Undo System**:
  - Transaction-like Job logging.
  - Undo support for Copy (delete), Move (move back), Delete (restore), Overwrite (restore backup).
- **TUI Actions**:
  - Shelf modes: Copy (`y`) and Move (`x`).
  - Execute Put (`p`) to target directory.
  - Execute Delete (`d`) to trash.
  - Global Undo (`u`).

### Fixed
- Browser startup: NOW shows files immediately (fixed wait-for-render issue).
- Dictionary navigation: Changed keybinding to `l`/`Right` (vim-style).

### Changed
- Startup policy: Always start in current working directory (`pwd`).

### Added (Sprint 1)
- **TUI Skeleton**: Basic layout with Browser, Shelf, Status bar using `gocui`.
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
