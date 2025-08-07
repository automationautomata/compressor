package utiles

import (
	"sync"

	"github.com/cheggaaa/pb/v3"
)

type Progress[I int64 | int] struct {
	open         bool
	progressChan chan I
	mu           sync.RWMutex
}

func NewProgress[I int64 | int](buffSize int) *Progress[I] {
	return &Progress[I]{true, make(chan I, buffSize), sync.RWMutex{}}
}

func (p *Progress[I]) Write(num I) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.open {
		p.progressChan <- num
		return true
	}
	return false
}

func (p *Progress[I]) IsOpen() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	isopen := p.open
	return isopen
}

func (p *Progress[I]) ShowProgress(total I) {
	if !p.open {
		return
	}
	switch any(total).(type) {
	case int64:
		go showProgress(int64(total), any(p.progressChan).(chan int64))
	case int:
		int64Chan := make(chan int64)
		go func() {
			defer close(int64Chan)
			for v := range p.progressChan {
				int64Chan <- int64(v)
			}
		}()
		go showProgress(int64(total), int64Chan)
	}
}

func showProgress(total int64, progressChan chan int64) {
	bar := pb.Start64(total)
	for incr := range progressChan {
		bar.Add64(incr)
		if bar.Current() >= total {
			break
		}
	}
	bar.Finish()
}

func (p *Progress[I]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.open {
		return
	}
	close(p.progressChan)
	p.open = false
}
