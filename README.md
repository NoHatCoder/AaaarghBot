# Installation

Install Golang.

Place the contents of this package in a folder that is neither in the GOROOT or GOPATH.

Install the `golang.org/x/net/context` package.

# Programs

There are several commands included under the `cmd` directory. All commands accept `-help` to list flags, but are otherwise minimally documented at present. Note that many flags are left over from Taktician and currently have no effect.

## cmd/playtak

A simple interface to play tak on the command line. To play black run:

```
go run main.go
```

To play white run:

```
go run main.go -white=human -black=minimax:5
```

## cmd/analyzetak

A program that reads PTN files and performs AI analysis on the terminal position.

```
analyzetak FILE.ptn
```

## cmd/taklogger

A bot that connects to playtak.com and logs all games it sees in PTN format.

## cmd/taktician

The AI driver for playtak.com.

Compile with:

```
go build
```

Can be used via:

```
taktician -user USERNAME -pass PASSWORD
```

Variables that control the behaviour can generally be found in the `cmd/taktician/main.go` and `cmd/taktician/taktician.go` files.

Only board sizes 5 and 6 are available. Depth is the primary method of adjusting strength, for reasonable performance generally don't go above depth 3 at size 6, and depth 4 at size 5. `t.ai.Diversify` sets the random component of evaluation, going below 100 runs a risk of making the AI too predictable, higher values makes the AI more random, and weaker.
