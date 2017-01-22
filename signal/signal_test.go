package signal

import (
  "testing"
  "sync"
  "time"
)

// Send signal when no listeners exist yet
func TestSend1(t *testing.T) {
  bc := NewBroadcast()
  bc.Send(&Signal{Key: "hello", Data: "hello world"})
  bc.Close()
}

func TestSend2(t *testing.T) {

  bc := NewBroadcast()
  var wa sync.WaitGroup

  for i := 0; i<10; i++ {
    wa.Add(1)
    go func(id int) {
      lis := bc.Join()
      wa.Done()
      //t.Logf("%v) Waiting...\n", id)
      signal, err := lis.Recv()
      if err != nil {
        t.Log(err)
      }
      t.Logf("%v) key: %v, data: %v\n", id, signal.Key, signal.Data)

    }(i)
  }

  wa.Wait()
  bc.Send(&Signal{Key: "hello", Data: "hello world"})
  bc.Close()
  t.Logf("Num.Listeners: %v", len(bc.listeners.m))
  time.Sleep(1*time.Second)
}

// Test send on closed channel

// Test receive on closed channel

// Test close broadcast twice

// Test leave a broadcast twice
