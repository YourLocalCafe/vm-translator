# vm-translator

A translator for a stack-based virtual machine language, written in Go. It reads one or more `.vm` files and generates Hack assembly (`.asm`) that implements the same program.

## Requirements

- Go (see `go.mod` for the version)
- No third-party dependencies

## Build

```sh
git clone https://github.com/YourLocalCafe/vm-translator.git
cd vm-translator
go build -o bin/vm-translator ./cmd
```

## Usage

Translate a single file:

```sh
./bin/vm-translator path/to/Main.vm
# writes path/to/Main.asm
```

Translate every `.vm` file in a directory into one combined program:

```sh
./bin/vm-translator path/to/Program/
# writes path/to/Program/Program.asm
```

Notes on input:

- File names must begin with an uppercase letter and end in `.vm` (e.g. `Main.vm`). In directory mode, other files are ignored.
- Directory paths must end with a trailing `/`.
- Whether the argument is treated as a file or a directory is decided by whether its last path segment contains a `.`.
- The program exits with status `1` and prints a message on any error.

## Output structure

Every generated `.asm` file has the same overall shape:

1. **Bootstrap** – sets `SP` to 261 and jumps to `Sys.init`.
2. **Translated code** – the commands from each input file, in order, each preceded by a comment naming the kind of command.
3. **Terminator** – an infinite loop (`(END)` / `@END` / `0;JMP`) so the program halts cleanly if it falls off the end.

Because the bootstrap always jumps to `Sys.init`, programs should define a `Sys.init` function.

## Supported commands

### Arithmetic and logic

| Command | Operation |
| ------- | --------- |
| `add`, `sub` | `x + y`, `x - y` |
| `neg` | `-y` |
| `eq`, `gt`, `lt` | Comparison; pushes `-1` (true) or `0` (false) |
| `and`, `or` | Bitwise `x & y`, `x \| y` |
| `not` | Bitwise `!y` |

### Memory access

```
push <segment> <index>
pop  <segment> <index>
```

| Segment | Maps to |
| ------- | ------- |
| `local` | `LCL + index` |
| `argument` | `ARG + index` |
| `this` | `THIS + index` |
| `that` | `THAT + index` |
| `temp` | `5 + index` |
| `pointer` | `0` → `THIS`, `1` → `THAT` |
| `static` | A per-file symbol `FileName.index` |
| `constant` | The literal value (`push` only) |

### Program flow

| Command | Description |
| ------- | ----------- |
| `label <name>` | Declares a label, scoped to the enclosing function |
| `goto <name>` | Unconditional jump |
| `if-goto <name>` | Pops the top of the stack and jumps if it is non-zero |

Labels are emitted as `Function$name`. Code that appears before any `function` command is scoped under `dummyFunctionWrapper`. The parser makes a first pass over each file to collect labels, so a `goto` or `if-goto` may refer to a label defined later in the same function. Jumping to an undefined label, or one that belongs to a different function, is reported as an error.

### Functions

| Command | Description |
| ------- | ----------- |
| `function <name> <nVars>` | Declares a function and initialises `nVars` local variables to 0 |
| `call <name> <nArgs>` | Pushes a return address and the caller's `LCL`, `ARG`, `THIS`, `THAT`, repositions `ARG` and `LCL`, then jumps to the function |
| `return` | Copies the return value to `ARG[0]`, restores the caller's frame and jumps back to the saved return address |

Return addresses are generated as `Function$ret.N`, where `N` is a running counter. Loop, comparison and skip labels use their own counters so they are unique across all files in a single run.

### Comments and whitespace

`//` comments (full-line or trailing) and blank lines are ignored, and leading/trailing spaces and tabs are trimmed.

## Example

`Main.vm`:

```
push constant 7
push constant 8
add
```

produces (after the bootstrap):

```asm
// Push constant command
@7
D=A
@SP
A=M
M=D
@SP
M=M+1
// Push constant command
@8
D=A
@SP
A=M
M=D
@SP
M=M+1
// Binary command
@SP
M=M-1
A=M
D=M
@SP
M=M-1
A=M
M=D+M
@SP
M=M+1
```

## Project layout

```
.
├── cmd/
│   └── main.go        Entry point: argument handling, file vs. directory mode, translation loop
├── parser/
│   └── parser.go      Reads commands, classifies them, extracts arguments, checks label scope
└── writer/
    └── writer.go      Assembly templates and the code generator for every command type
```

| Package | Responsibility |
| ------- | -------------- |
| `cmd` | Decides what to translate, creates the writer and feeds it each parsed command |
| `parser` | `Advance`, `CommandType`, `Arg1`, `Arg2`; tracks the current function and the label map |
| `writer` | `WriteArithmetic`, `WritePushPop`, `WriteLabel`, `WriteGoto`, `WriteIf`, `WriteFunction`, `WriteCall`, `WriteReturn` |
