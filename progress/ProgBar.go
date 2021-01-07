package progress

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Bar struct {
	Total   int
	current int
	last    int
	tick    *time.Ticker
}

const (
	maxbars  int = 100
	interval     = 300 * time.Millisecond
)

func (p *Bar) print() {
	p.last = p.current
	bars := p.calcBars(p.current) //算长度
	spaces := maxbars - bars - 1
	percent := 100 * (float32(p.current) / float32(p.Total))
	builder := strings.Builder{}
	for i := 0; i < bars; i++ {
		builder.WriteRune('=')
	}
	builder.WriteRune('>')
	for i := 0; i <= spaces; i++ {
		builder.WriteRune(' ')
	}
	fmt.Printf(" \r[%s] %3.2f%% (%d/%d)", builder.String(), percent, p.current, p.Total)
}

func (p *Bar) printComplete() {
	p.print()
	fmt.Print("\n")
}

func (p *Bar) calcBars(portion int) int {
	if portion == 0 {
		return portion
	}
	return int(float32(maxbars) / (float32(p.Total) / float32(portion)))
}

func (p *Bar) Run() {
	go func() {
		p.tick = time.NewTicker(interval) //定时发送tick tock
		for range p.tick.C {
			if p.last != p.current {
				log.Println("Ding")
				p.print() //如果收到tick tock，就打印状态
			}
			if p.current >= p.Total {
				log.Println("Dong")
				p.printComplete()
				p.tick.Stop()
				return
			}
		}
	}()
}

func (p *Bar) Current(v int) {
	p.current = v
}
func (p *Bar) Add(v int) {
	p.current += v
}
