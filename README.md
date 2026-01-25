# heatPlot

`heatPlot` is a Go tool for generating animated heatmaps (GIFs) from mathematical functions. It parses mathematical expressions involving variables `x`, `y`, and `t` (time), and renders them as a heatmap.

## Installation

Ensure you have Go installed (Go 1.22+ recommended).

Clone the repository and install the dependencies:

```bash
git clone https://bitbucket.org/arran4/heatplot.git
cd heatplot
go mod tidy
```

## Usage

There are three main commands available in the `cmd` directory:

### 1. heatPlot

Generates a heatmap GIF from a provided mathematical formula.

**Build:**

```bash
go build -o heatPlot cmd/heatPlot/main.go
```

**Run:**

```bash
./heatPlot [flags] "formula"
```

**Flags:**

- `-hcc`: Heat colour count (default 126).
- `-speed`: Duration between frames (default 100ms).
- `-pointSize`: Scale of x/y steps (default 0.1).
- `-scale`: Magnification (default 2).
- `-tlb`: Time lower bound (start T, default 0).
- `-tub`: Time upper bound (end T, default 100).
- `-size`: Cartesian plane size (default 100, i.e., -100 to 100).
- `-outputFile`: Output filename (default "./out.gif").
- `-footerText`: Footer text (default "http://github.com/arran4/").

**Example:**

```bash
./heatPlot -outputFile="example.gif" "y = x * sin(t/10)"
```

### 2. heatPlotRandom

Generates random functions and renders one that meets certain "interestingness" criteria (complexity, movement).

**Build:**

```bash
go build -o heatPlotRandom cmd/heatPlotRandom/main.go
```

**Run:**

```bash
./heatPlotRandom [flags]
```

**Flags:**

Similar to `heatPlot`, with additional criteria for random generation.

### 3. whatFunctions

Lists all available single and double parameter functions supported by the parser.

**Build:**

```bash
go build -o whatFunctions cmd/whatFunctions/main.go
```

**Run:**

```bash
./whatFunctions
```

## Formula Syntax

The parser supports:
- Variables: `x`, `y`, `t`
- Constants: Numbers
- Operators: `+`, `-`, `*`, `/`, `%` (modulus), `^` (power)
- Functions: `sin`, `cos`, `tan`, `abs`, `max`, `min`, `pow`, etc. (See `whatFunctions` for full list)
- Grouping: `()`

Example Formulas:
- `y = x + t`
- `y / 4 = x * sin(t)`
- `val = x^2 + y^2` (Note: The parser expects an assignment, usually `something = something else`, but internally evaluates `RHS - LHS` for the heatmap value).

## Development

### Prerequisites

- Go 1.22 or later
- Make

### Building

To regenerate the parser from `calc.y`:

```bash
make setup
make yacc
```

### Testing

Run the tests:

```bash
go test ./...
```
