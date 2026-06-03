# cccat - My own version of `cat`

A clone of the Unix `cat` tool, written in Go. Part of the
[Coding Challenges](https://codingchallenges.fyi/challenges/challenge-cat) series.
 
This is a **dev log**, not a manual. I want to share the design choices I made,
the bugs that taught me something, and the Go ideas I learned while rebuilding
a tool I had used many times without thinking about how it works.
 
---
 
## What it does
 
```bash
cccat file.txt              # print a file
cccat -n file.txt           # number all lines
cccat -b file.txt           # number non-blank lines only
cccat file1.txt file2.txt   # print multiple files
echo "hi" | cccat -n        # read from stdin
```
 
---
 
## Things I had to learn first
 
### A simple mental model: water through pipes
 
Think of a file as a **water tank** and a program as a **hose**. When you run
`cat file.txt`, the hose moves water from the tank to your terminal. The water
flows in small pieces — you don't move all the water first and then pour it out.
This is why `cat` can print a 50 GB file on a computer with 8 GB of RAM.
 
The Unix `|` operator is called a **pipe** on purpose. Kernighan and Pike wrote
that Unix's power comes from the way small tools connect through pipes, "because
that is a universal interface."[^upe]
 
### stdin, stdout, stderr
 
Every Unix process starts with three pipes open. Kerrisk explains them as file
descriptors 0, 1, and 2:[^tlpi]
 
- **stdin (0)** — where the program reads input. Keyboard by default, or a pipe.
- **stdout (1)** — where the program writes output. Terminal by default.
- **stderr (2)** — a separate channel for error messages.
In Go these are `os.Stdin`, `os.Stdout`, and `os.Stderr`. They all use the same
`io.Reader` and `io.Writer` interfaces as regular files. That is why a file and
stdin can share the same code.
 
### Streaming I/O in Go
 
Donovan and Kernighan show these two interfaces early in *The Go Programming
Language* because they are the base of almost all I/O in Go:[^gopl]
 
```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```
 
Files, sockets, pipes, stdin, stdout — they all use these. You can connect any
source to any destination with one call: `io.Copy(dst, src)`. That is the hose,
in code.
 
---
 
## Design decisions
 
**Input as an interface.** `FileInput` and `StdinInput` both implement `Open()
(io.ReadCloser, error)`. The main loop never asks which one it has. It just
opens and copies.
 
**Decorator for display.** `-n` and `-b` change *how* content is displayed, not
*where* it comes from. I modeled this as a `CopyFunc` type. The flag parser
picks one of three versions (`copyPlain`, `copyWithLineNumbers`,
`copyWithNonBlankLineNumbers`). The loop knows nothing about flags.
 
**`flag` package over manual parsing.** I parsed arguments by hand first to
understand the rules `cat` follows. Then I switched to the standard `flag`
package, which follows the same rules.
 
**One file.** At ~150 lines, splitting into packages would be extra work without
real benefit. Knowing when *not* to add structure is also a decision.
 
---
 
## Bugs that taught me something
 
### 1. `string(int)` is not what you think
 
```go
builder.WriteString(string(number)) // prints a Unicode character, not a number
```
 
In Go, `string(65)` is `"A"`, not `"65"`. Use `strconv.Itoa(n)` or `%d`.
The compiler has warned about this since Go 1.15.
 
### 2. `strings.Builder` declared outside a loop accumulates
 
I never reset the builder between iterations. Line 2 contained line 1 + line 2.
The fix: I didn't need the builder. `fmt.Fprintf(dst, "%6d\t%s", n, line)`
writes directly to the output.
 
### 3. `defer` runs at function exit, not loop iteration
 
`defer reader.Close()` inside a loop keeps all files open until `main` ends.
The fix: move the logic to a `processInput` function so `defer` fires at the
end of each file's processing, not at the end of the program.
 
### 4. `bufio.Scanner` hides the trailing newline
 
`Scanner` removes the `\n` delimiter, so I always added one back. But the last
line of a file may not have a `\n`. I was adding a byte that wasn't there.
 
Fix: switch to `bufio.Reader.ReadString('\n')`, which keeps the delimiter. Note
that it can return data *and* an error at the same time — process the line
before checking the error.
 
### 5. Filter vs. treat differently (the `-b` bug)
 
I used `continue` on blank lines, so they disappeared from output. But `cat -b`
*prints* every line — it just doesn't *number* the blank ones.
 
```go
if line == "\n" {
    fmt.Fprint(dst, line)                      // print, don't number
} else {
    fmt.Fprintf(dst, "%6d\t%s", number, line)
    number++
}
```
 
**Lesson:** ask whether a condition changes *what* is shown or *how* it is
shown. These are different things.
 
---
 
## Known limitations
 
- Combined flags like `-nb` don't work. Go's `flag` package treats `-nb` as one
  flag named `nb`.
- Only `-n` and `-b` are implemented.
---
 
## Running it
 
```bash
go build -o cccat
./cccat -n file.txt
```
 
---
 
## Further reading
 
- **Donovan & Kernighan (2015).** *The Go Programming Language.* Addison-Wesley.
- **Kerrisk (2010).** *The Linux Programming Interface.* No Starch Press.
- **Kernighan & Pike (1984).** *The Unix Programming Environment.* Prentice Hall.
- **Stevens & Rago (2013).** *Advanced Programming in the UNIX Environment* (3rd ed.). Addison-Wesley.
[^upe]: Kernighan & Pike (1984). *The Unix Programming Environment.* Prentice Hall.
[^tlpi]: Kerrisk (2010). *The Linux Programming Interface.* No Starch Press. See the chapter on file descriptors, around p. 73.
[^gopl]: Donovan & Kernighan (2015). *The Go Programming Language.* Addison-Wesley. `io.Reader` and `io.Writer` are introduced in Chapter 7.