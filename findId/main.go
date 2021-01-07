package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func Equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func contains(s []rune, e rune) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println(time.Now())
	file, _ := os.Open("the stupid large file")
	defer file.Close()
	resultFile, _ := os.OpenFile("result.txt", os.O_CREATE|os.O_WRONLY, 0644)
	defer resultFile.Close()
	datawriter := bufio.NewWriter(resultFile)
	defer datawriter.Flush()

	r := bufio.NewReader(file)
	customerId := []rune("customerId")
	var typeStack []rune
	var idStack []rune
	var mode = 0
	var i = 0
	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Println(err)
				break
			}
		} else {
			switch c {
			case '=':
				if Equal(typeStack, customerId) {
					mode = 1
					typeStack = typeStack[:0]
				} else {
					typeStack = typeStack[:0]
				}
			case ',':
				if mode == 1 {
					datawriter.WriteString(string(idStack) + "\n")
					if i < 1000 {
						i++
					} else {
						datawriter.Flush()
						i = 0
					}
					typeStack = typeStack[:0]
					idStack = idStack[:0]
					mode = 0
				} else {
					typeStack = typeStack[:0]
				}
			default:
				if mode == 1 {
					idStack = append(idStack, c)
				} else {
					if c != 'c' && c != 'u' && c != 's' && c != 't' && c != 'o' && c != 'm' && c != 'e' && c != 'r' && c != 'I' && c != 'd' {
						typeStack = typeStack[:0]
					} else {
						typeStack = append(typeStack, c)
					}
				}
			}
		}
	}
	fmt.Println(time.Now())
}
