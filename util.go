package main

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

type Mass struct {
	batchSize int
	count     int
	step      int
	start     int
	end       int
}

func New(count, batchSize int) *Mass {
	mass := new(Mass)

	if count < 0 {
		count = 0
	}

	if batchSize <= 0 {
		batchSize = 1
	}

	mass.count = count
	mass.batchSize = batchSize
	return mass
}

func (m *Mass) Iter() bool {
	roundSize := m.batchSize

	if m.step != 0 {
		m.start += m.batchSize
	}

	if m.start+roundSize > m.count {
		roundSize = m.count - m.start
	}

	m.end = m.start + roundSize

	m.step++

	return m.start < m.count
}

func (m *Mass) Begin() int {
	return m.start
}

func (m *Mass) End() int {
	return m.end
}

type BatchRunner struct {
	Total     int
	BatchSize int
	Worker    chan struct{}
	Mass      *Mass
	Wg        sync.WaitGroup
}

func NewBatchRunner(total, batch, worker int) *BatchRunner {
	return &BatchRunner{
		Total:     total,
		BatchSize: batch,
		Worker:    make(chan struct{}, worker),
		Mass:      New(total, batch),
	}
}

func (r *BatchRunner) Run(f func()) {
	select {
	case r.Worker <- struct{}{}:
		ants.Submit(func() {
			defer func() {
				<-r.Worker
				r.Wg.Done()
			}()
			f()
		})
	}
}

func (r *BatchRunner) Iter() bool {
	t := r.Mass.Iter()
	if t {
		r.Wg.Add(1)
	} else {
		r.Wg.Wait()
	}
	return t
}

func (r *BatchRunner) Begin() int {
	return r.Mass.Begin()
}

func (r *BatchRunner) End() int {
	return r.Mass.End()
}
