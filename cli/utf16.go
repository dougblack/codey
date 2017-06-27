package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"io"
	"os"
)

type Utf16Command struct {
	utf16 bool
}

func (*Utf16Command) Name() string     { return "utf16" }
func (*Utf16Command) Synopsis() string { return "decode file as utf16" }
func (*Utf16Command) Usage() string {
	return `codey utf16 [file]:
	decode file as utf16
`
}

func (u *Utf16Command) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&u.utf16, "utf16", false, "decode file as utf16")
}

const (
	Normal      = 0
	Surrogate   = 1
	Utf16Ignore = -1
)

func classify16(unit uint16) int {
	if unit >= 0xd800 && unit <= 0xdfff {
		return Surrogate
	} else {
		return Normal
	}
}

func decode16(encoded []uint16) (character rune) {
	if len(encoded) == 0 {
		return 0
	}

	class := classify16(encoded[0])

	if class == Normal {
		character = rune(encoded[0])
	} else if class == Surrogate {
		high := rune(encoded[0])
		low := rune(encoded[1])
		character = rune(0x10000 + (high-0xd800)*0x400 + (low - 0xdc00))
	}
	return character
}

func (u *Utf16Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() != 1 {
		fmt.Printf("Expected exactly one argument for file to decode")
		return subcommands.ExitUsageError
	}
	parse16(f.Arg(0))
	return subcommands.ExitSuccess
}

func parse16(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	data := make([]byte, 4096)
	for {
		data = data[:cap(data)]
		n, err := f.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
			return
		}
		data = data[:n]

		var encoded []uint16
		var class int
		var unit uint16
		for i := 0; i < len(data); i += 2 {
			unit = uint16(data[i+1]) | uint16(data[i])<<8
			class = classify16(unit)
			if class == Utf16Ignore {
				encoded = []uint16{}
			} else {
				encoded = append(encoded, unit)
				if class == Surrogate {
					pair := uint16(data[i+3]) | uint16(data[i+2])<<8
					encoded = append(encoded, pair)
					i += 2
				}
				if len(encoded) > 0 {
					var byteBuffer bytes.Buffer
					for x := 0; x < len(encoded); x++ {
						byteBuffer.WriteString(fmt.Sprintf("%08b ", encoded[x]))
					}
					fmt.Printf("%34s", byteBuffer.String())
					decoded := decode16(encoded)
					fmt.Printf("-> U+%X(%q)\n", decoded, decoded)
				}
				encoded = []uint16{}
			}
		}
	}
}
