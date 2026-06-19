# Agents Guidelines

## UI Testing (`test-ui` mode)
The `heatPlot` tool features a `test-ui` mode to headlessly verify the UI logic and output without launching a window. This allows programmatic testing of the application state.

### Usage
Run the tool with the `-test-ui` and `-test-ui-out` flags:
```bash
./heatPlot -test-ui=events.json -test-ui-out=output.png
```

### JSON Input Format
The `events.json` file should contain an array of key events to simulate user interaction. Each event can specify a `Rune` (a single character string) or a `Code` (the string representation of a `key.Code` such as `"CodeReturnEnter"`, `"CodeEscape"`, `"CodeLeftArrow"`, `"CodeRightArrow"`, `"CodeF1"`, etc.).

Example `events.json`:
```json
[
  {"Code": "CodeRightArrow"},
  {"Rune": "t"},
  {"Rune": "y"},
  {"Rune": " "},
  {"Rune": "="},
  {"Rune": " "},
  {"Rune": "x"},
  {"Code": "CodeReturnEnter"}
]
```

This will apply the events sequentially to the UI state and output the final rendered image to the path specified by `-test-ui-out`.
