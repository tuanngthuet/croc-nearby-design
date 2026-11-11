package croc

import (
	"strconv"
	"testing"

	"github.com/schollz/croc/v11/src/utils"
)

// moreRelayPortsThanFindOpenPortsCanScan puts setupLocalRelay's "not enough
// open ports" branch within reach without occupying real ports: FindOpenPorts
// scans at most 200 ports, so a longer port list can never be satisfied.
func moreRelayPortsThanFindOpenPortsCanScan() []string {
	ports := make([]string, 250)
	for i := range ports {
		ports[i] = strconv.Itoa(9009 + i)
	}
	return ports
}

func TestSetupLocalRelayReturnsErrorInLibraryMode(t *testing.T) {
	c := &Client{Options: Options{
		RelayPorts:  moreRelayPortsThanFindOpenPortsCanScan(),
		LibraryMode: true,
	}}
	if err := c.setupLocalRelay(make(chan error, 2)); err == nil {
		t.Fatal("expected an error instead of a panic")
	}
}

func TestSetupLocalRelayPanicsWithoutLibraryMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic outside library mode")
		}
	}()
	c := &Client{Options: Options{RelayPorts: moreRelayPortsThanFindOpenPortsCanScan()}}
	_ = c.setupLocalRelay(make(chan error, 2))
}

func TestUploadLimitReturnsErrorInLibraryMode(t *testing.T) {
	_, err := New(Options{
		SharedSecret:   utils.GetRandomName(),
		IsSender:       true,
		ThrottleUpload: "abc",
		LibraryMode:    true,
	})
	if err == nil {
		t.Fatal("expected an error instead of a panic")
	}
}

func TestUploadLimitPanicsWithoutLibraryMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic outside library mode")
		}
	}()
	_, _ = New(Options{
		SharedSecret:   utils.GetRandomName(),
		IsSender:       true,
		ThrottleUpload: "abc",
	})
}
