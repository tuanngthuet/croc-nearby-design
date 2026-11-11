package croc

// TransferLifecycle is the exported twin of transferLifecycle, so an embedder
// can tell "still waiting for the peer" (ChannelSecured false) from a transfer
// that is actually moving data.
type TransferLifecycle = transferLifecycle

// LifecycleSnapshot returns a consistent copy of the handshake step flags. It
// is safe to call from any goroutine, at any rate.
func (c *Client) LifecycleSnapshot() TransferLifecycle {
	return c.lifecycleSnapshot()
}
