# demo/demo.tape
#
# VHS script (https://github.com/charmbracelet/vhs) to record the
# catnet-tui demonstration GIF used in README.md.
#
# How to generate:
#   1. go install github.com/charmbracelet/vhs@latest   (requires ttyd + ffmpeg in PATH)
#   2. go build -o demo/catnet-tui ./cmd/catnet-tui
#   3. vhs demo/demo.tape
#   4. Review demo/demo.gif and commit (do not track binary demo/catnet-tui)
#
# IMPORTANT — do not run as an automated CI step:
# scan uses the real network (ARP/ICMP). Hosted runners (GitHub Actions)
# lack visibility into a real LAN, so output would be empty, slow, or flaky.
# Generate this GIF locally on your own network and re-run manually when
# UI changes significantly. Keep .tape versioned; .gif is the artifact.

Output demo/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Dracula"
Set TypingSpeed 60ms
Set Padding 20

# --- Target input screen ---
Type "./demo/catnet-tui"
Enter
Sleep 800ms

# Adjust this range to a small, known range of your local network
# before recording — avoids exposing unnecessary real topology and keeps
# the scan fast and predictable in the GIF.
Type "192.168.100.1-20"
Sleep 300ms
Enter

# --- Wait for scan completion instead of fixed sleep (more robust) ---
Wait+Screen@15s /Scan completed/

Sleep 1s

# --- Navigate results table ---
Down@250ms 3
Sleep 500ms
Up@250ms 3
Sleep 500ms

# --- Toggle topology map (key feature) ---
Type "t"
Sleep 2s

# --- Toggle back to table + map ---
Type "t"
Sleep 1500ms

# --- Exit ---
Type "q"
Sleep 500ms
