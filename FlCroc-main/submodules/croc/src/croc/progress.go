package croc

import (
	"path/filepath"
)

// TransferProgress is a consistent snapshot of the transfer counters.
//
// The same values are also exported as plain fields on Client, but reading
// them from another goroutine is not safe: croc updates them from its transfer
// goroutines without any lock an embedder can share. FilesHasFinished is a map,
// and a concurrent map read is a fatal runtime error that cannot be recovered.
// ProgressSnapshot takes progressMu and returns a copy.
type TransferProgress struct {
	// TotalFiles is the number of entries in the transfer manifest.
	TotalFiles int
	// TotalSize is the sum of every file size in the manifest.
	TotalSize int64
	// TransferredSize is the bytes moved so far, including the files that
	// already finished.
	TransferredSize int64
	// CurrentFile is the base name of the file in flight.
	CurrentFile string
	// CurrentFileIndex is 1-based; 0 means the manifest is not known yet.
	CurrentFileIndex int
	// FinishedFiles is the number of files croc marked complete.
	FinishedFiles int
}

// resetTransferCounters zeroes the counters for the file about to be moved.
func (c *Client) resetTransferCounters() {
	c.progressMu.Lock()
	c.TotalSent = 0
	c.TotalChunksTransferred = 0
	c.progressMu.Unlock()
}

func (c *Client) setTotalSent(v int64) {
	c.progressMu.Lock()
	c.TotalSent = v
	c.progressMu.Unlock()
}

func (c *Client) addTotalSent(n int64) {
	c.progressMu.Lock()
	c.TotalSent += n
	c.progressMu.Unlock()
}

func (c *Client) setCurrentFileNum(n int) {
	c.progressMu.Lock()
	c.FilesToTransferCurrentNum = n
	c.progressMu.Unlock()
}

func (c *Client) setFilesToTransfer(files []FileInfo) {
	c.progressMu.Lock()
	c.FilesToTransfer = files
	c.progressMu.Unlock()
}

func (c *Client) markFileFinished() {
	c.progressMu.Lock()
	c.FilesHasFinished[c.FilesToTransferCurrentNum] = struct{}{}
	c.progressMu.Unlock()
}

func (c *Client) hasFileFinished(i int) bool {
	c.progressMu.RLock()
	_, ok := c.FilesHasFinished[i]
	c.progressMu.RUnlock()
	return ok
}

// recordReceivedChunk adds n bytes to the counters and reports whether the
// current file is complete. chunkCount is passed in because the caller holds
// receiveMutex, which is what keeps CurrentFileChunkCount stable.
func (c *Client) recordReceivedChunk(n int64, chunkCount int) (finished bool) {
	c.progressMu.Lock()
	defer c.progressMu.Unlock()
	c.TotalSent += n
	c.TotalChunksTransferred++
	if c.TotalChunksTransferred == chunkCount {
		return true
	}
	i := c.FilesToTransferCurrentNum
	return i >= 0 && i < len(c.FilesToTransfer) && c.TotalSent == c.FilesToTransfer[i].Size
}

// ProgressSnapshot returns a consistent copy of the transfer counters. It is
// safe to call from any goroutine and at any rate.
func (c *Client) ProgressSnapshot() TransferProgress {
	c.progressMu.RLock()
	defer c.progressMu.RUnlock()

	files := c.FilesToTransfer
	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	i := c.FilesToTransferCurrentNum
	var done int64
	for j := 0; j < i && j < len(files); j++ {
		done += files[j].Size
	}

	current := ""
	if i >= 0 && i < len(files) {
		current = filepath.Base(files[i].Name)
	}

	return TransferProgress{
		TotalFiles:       len(files),
		TotalSize:        totalSize,
		TransferredSize:  done + c.TotalSent,
		CurrentFile:      current,
		CurrentFileIndex: i + 1,
		FinishedFiles:    len(c.FilesHasFinished),
	}
}
