# lazycd

**lazycd** is a terminal UI (TUI) file manager inspired by macOS utilities like "Dropover" or "Yoink". It introduces the concept of a "Shelf" — a temporary holding area for files and directories — directly into your command-line workflow.

Instead of wrestling with complex `cp` or `mv` commands involving long paths, `lazycd` allows you to pick up files from multiple locations, place them on a virtual shelf, navigate to your destination, and drop them all at once.

## Features

- **TUI Navigation**: Fast, keyboard-centric file system browser using `gocui`.
- **The Shelf**: A persistent temporary storage area. Add files from anywhere, then decide what to do with them later.
- **Visual Feedback**: Clear distinct views for the file browser and your shelf items.
- **Conflict Resolution**: Smart handling of conflicting file names (Skip, Rename, Overwrite).
- **Undo Support**: Accidentally moved a file? Undo the last operation with a single keystroke.
- **Bulk Operations**: Select multiple files, add them to the shelf, and process them in batches.

## Installation

### Build from Source

Requirements: Go 1.20+

```bash
# Clone the repository
git clone https://github.com/jangmyun/lazycd.git
cd lazycd

# Build the binary
go build -o lazycd ./cmd/lazycd

# Move to a location in your $PATH
mv lazycd /usr/local/bin/
```

## Usage

Run the application:

```bash
lazycd
```

The interface is divided into two main panels:
1. **Left Panel (Browser)**: Navigate your file system.
2. **Right Panel (Shelf)**: View and manage items you've picked up.

### Keybindings

#### Global
| Key | Action |
| --- | --- |
| `Tab` | Switch focus between Browser and Shelf |
| `u` | Undo last operation |
| `?` | Toggle Help overlay |
| `q` / `Ctrl+c` | Quit application |

#### Browser Panel
| Key | Action |
| --- | --- |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `l` / `→` | Enter directory |
| `h` / `←` | Go to parent directory |
| `Space` | Toggle selection (multi-select) |
| `.` | Toggle hidden files |
| `Enter` | Toggle file details view |
| `a` | **Add** selected/current item to Shelf |
| `t` | Set current directory as **Target** for operations |

#### Shelf Panel
| Key | Action |
| --- | --- |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Space` | Toggle selection for batch operations |
| `y` | Set mode to **Copy** (default) |
| `x` | Set mode to **Move** |
| `r` | **Remove** item from Shelf (does not delete file) |
| `d` | **Delete** file permanently (moves to trash) |
| `p` | **Put** items to Target directory |

### Example Workflow: Moving Files

1. Navigate to the source directory in the **Browser**.
2. Select files using `Space`.
3. Press `a` to add them to the **Shelf**.
4. Navigate to the destination directory.
5. Press `t` to set the current directory as the **Target**.
6. Press `Tab` to switch to the **Shelf**.
7. Press `x` to set the operation mode to **Move**.
8. Press `p` to put (move) the files from the Shelf to the Target directory.
