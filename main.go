package main

import (
	"fmt"
	"os"
	"io"
	"bytes"
)

const (
	One = 0
	Two = 1
	Three = 2
	Four = 3
	Cont = 4
	Ignore = -1
	Invalid = -2
)

func classify(b byte) int {
	if b == 0 {
		return Ignore
	} else if b >> 7 == 0 {
		return One
	} else if b >> 5 == 6 {
		return Two
	} else if b >> 4 == 14 {
		return Three
	} else if b >> 3 == 30 {
		return Four
	} else if b >> 6 == 2 {
		return Cont
	}
	return Invalid
}

func decode(encoded []byte) (character rune) {
	if len(encoded) == 0 {
		return 0
	}

	class := classify(encoded[0])

	if class == One {
		character = rune(encoded[0])
	} else if class == Two {
		character = rune(encoded[0] & 31) << 6 |
					rune(encoded[3] & 63)
	} else if class == Three {
		character = rune(encoded[0] & 15) << 12 |
					rune(encoded[2] & 63) << 6 |
					rune(encoded[3] & 63)
	} else if class == Four {
		character = rune(encoded[0] & 7) << 18 |
					rune(encoded[1] & 63) << 12 |
					rune(encoded[2] & 63) << 6 |
					rune(encoded[3] & 63)
	}
	return character
}

func main() {
	f, err := os.Open(os.Args[1])
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

		var encoded []byte
		var class int
		for i := 0; i < len(data); i++ {
			b := data[i]
			class = classify(b)
			if class == Ignore {
				encoded = []byte{}
			} else if class == Cont {
				encoded = append(encoded, b)
			} else if class == Invalid {
				fmt.Printf("Invalid byte: %b", b)
				encoded = []byte{}
			} else {
				var j int
				for j = 0; j <= class; j++ {
					encoded = append(encoded, data[i+j])
				}
				i += j - 1
				if len(encoded) > 0 {
					var byteBuffer bytes.Buffer
					for x := 0; x < len(encoded); x++ {
						byteBuffer.WriteString(fmt.Sprintf("%b ", encoded[x]))
					}
					fmt.Printf("%36s", byteBuffer.String())
					decoded := decode(encoded)
					fmt.Printf(" -> U+%X(%q)\n", decoded, decoded)
				}
				encoded = []byte{}
			}
		}
	}
}
