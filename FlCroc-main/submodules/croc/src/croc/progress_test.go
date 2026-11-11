package croc

import (
	"sync"
	"testing"
)

func TestProgressSnapshotValues(t *testing.T) {
	c := &Client{
		FilesToTransfer: []FileInfo{
			{Name: "/a/one.txt", Size: 10},
			{Name: "/a/two.txt", Size: 30},
		},
		FilesHasFinished: make(map[int]struct{}),
	}
	c.setCurrentFileNum(1)
	c.setTotalSent(5)

	got := c.ProgressSnapshot()
	if got.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", got.TotalFiles)
	}
	if got.TotalSize != 40 {
		t.Errorf("TotalSize = %d, want 40", got.TotalSize)
	}
	// TotalSent counts the current file only, so the finished one is added on
	// top: 10 (one.txt) + 5 (partial two.txt).
	if got.TransferredSize != 15 {
		t.Errorf("TransferredSize = %d, want 15", got.TransferredSize)
	}
	if got.CurrentFile != "two.txt" {
		t.Errorf("CurrentFile = %q, want %q", got.CurrentFile, "two.txt")
	}
	if got.CurrentFileIndex != 2 {
		t.Errorf("CurrentFileIndex = %d, want 2", got.CurrentFileIndex)
	}
	if got.FinishedFiles != 0 {
		t.Errorf("FinishedFiles = %d, want 0", got.FinishedFiles)
	}

	c.markFileFinished()
	if got := c.ProgressSnapshot(); got.FinishedFiles != 1 {
		t.Errorf("FinishedFiles = %d, want 1", got.FinishedFiles)
	}
}

func TestRecordReceivedChunk(t *testing.T) {
	c := &Client{
		FilesToTransfer: []FileInfo{{Name: "one.txt", Size: 10}},
		FilesHasFinished: make(map[int]struct{}),
	}

	if finished := c.recordReceivedChunk(4, 2); finished {
		t.Error("recordReceivedChunk(4, 2) = true, want false")
	}
	if finished := c.recordReceivedChunk(4, 2); !finished {
		t.Error("second recordReceivedChunk(4, 2) = false, want true (chunk count reached)")
	}

	c.resetTransferCounters()
	if finished := c.recordReceivedChunk(10, 8); !finished {
		t.Error("recordReceivedChunk(10, 8) = false, want true (size reached)")
	}
}

// The counters live on exported Client fields, so an embedder can read them
// from its own goroutine. Run this with -race: before progressMu existed, the
// map access in particular was a fatal error rather than a recoverable panic.
func TestProgressSnapshotConcurrentAccess(t *testing.T) {
	files := []FileInfo{{Name: "one.txt", Size: 10}, {Name: "two.txt", Size: 10}}
	c := &Client{
		FilesToTransfer:  files,
		FilesHasFinished: make(map[int]struct{}),
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				c.addTotalSent(1)
				c.setCurrentFileNum(i % 2)
				c.markFileFinished()
				_ = c.hasFileFinished(i % 2)
				_ = c.recordReceivedChunk(1, 0)
			}
		}(i)
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_ = c.ProgressSnapshot()
			}
		}()
	}

	for n := 0; n < 20000; n++ {
		_ = c.ProgressSnapshot()
	}

	close(stop)
	wg.Wait()
}
