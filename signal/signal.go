package signal

import (
  "errors"
  "sync"
)

var (
  ErrListenerClosed = errors.New("listener closed")
)

type Signal struct {
  Key string
  Data interface{}
}

type listenersMap struct {
  sync.RWMutex
	m     map[*Listener]*Listener
}

func NewBroadcast() *Broadcast {

  bc := &Broadcast{
    send_ch: make(chan *Signal),
    close_ch: make(chan bool),
    listeners: listenersMap{m: make(map[*Listener]*Listener)},
    cond: sync.Cond{L: &sync.Mutex{}},
  }

  go bc.run()

  return bc
}

type Broadcast struct {
  send_ch chan *Signal
  close_ch chan bool
  listeners listenersMap
  closed bool
  cond sync.Cond
}

func (bc *Broadcast) run() {

  defer func() {
    close(bc.send_ch)
    close(bc.close_ch)
    bc.cond.L.Lock()
    bc.closed = true
    bc.cond.L.Unlock()
    bc.cond.Signal()
  }()

  exit := false

  for !exit {
    select {

    case signal := <-bc.send_ch:
      func () {
        defer bc.listeners.RUnlock()
        bc.listeners.RLock()
        for _, listener := range bc.listeners.m {
          listener.recv_ch <- signal
        }
      }()

    case <- bc.close_ch:
      for _, listener := range bc.listeners.m {
        listener.Leave()
      }
      exit = true
    }
  }
}

func (bc *Broadcast) Send(signal *Signal) {
  bc.send_ch <- signal
}

func (bc *Broadcast) Join() *Listener {

  listener := &Listener{
    recv_ch: make(chan *Signal),
    bc: bc,
  }

  defer bc.listeners.Unlock()
  bc.listeners.Lock()
  bc.listeners.m[listener] = listener
  return listener
}

func (bc *Broadcast) removeListener(listener *Listener) {
  defer bc.listeners.Unlock()
  bc.listeners.Lock()
  close(listener.recv_ch)
  delete(bc.listeners.m, listener)
}

func (bc *Broadcast) Close() {
  defer bc.cond.L.Unlock()
  bc.cond.L.Lock()
  if !bc.closed {
    bc.close_ch <- true
    bc.cond.Wait()
  }
}

// A listener must be used only from a goroutine.
type Listener struct {
  recv_ch chan *Signal
  close_ch chan bool
  bc *Broadcast
  closed bool
}

// Must be called from the same goroutine over all the lifecycle of the listener
func (l *Listener) Recv() (*Signal, error)  {

  if l.closed {
    return nil, ErrListenerClosed
  }

  select {
  case signal := <-l.recv_ch:
    return signal, nil
  case <- l.close_ch:
    l.closed = true
    return nil, ErrListenerClosed
  }
}

// Can be called from another thread
func (l *Listener) Leave() {
  go func() {
    l.close_ch <- true
    close(l.close_ch)
  }()
  l.bc.removeListener(l)
}
