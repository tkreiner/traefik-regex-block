package traefik_regex_block

import (
	"net"
	"sync"
	"testing"
	"time"
)

// TestArrayStorage_Block validates blocking functionality
func TestArrayStorage_Block(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	ip := net.ParseIP("192.168.1.1")
	err := storage.Block(ip, 1)

	if err != nil {
		t.Errorf("Block should not return error: %v", err)
	}

	if !storage.IsBlocked(ip) {
		t.Error("IP should be blocked after Block call")
	}
}

// TestArrayStorage_BlockNilIP validates handling of nil IP
func TestArrayStorage_BlockNilIP(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	err := storage.Block(nil, 1)

	if err != nil {
		t.Errorf("Block with nil IP should not return error: %v", err)
	}
}

// TestArrayStorage_IsBlockedNilIP validates handling of nil IP in IsBlocked
func TestArrayStorage_IsBlockedNilIP(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	blocked := storage.IsBlocked(nil)

	if blocked {
		t.Error("nil IP should not be considered blocked")
	}
}

// TestArrayStorage_UnBlockNilIP validates handling of nil IP in UnBlock
func TestArrayStorage_UnBlockNilIP(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	err := storage.UnBlock(nil)

	if err != nil {
		t.Errorf("UnBlock with nil IP should not return error: %v", err)
	}
}

// TestArrayStorage_IsBlocked validates block checking
func TestArrayStorage_IsBlocked(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	ip := net.ParseIP("192.168.1.1")

	// Should not be blocked initially
	if storage.IsBlocked(ip) {
		t.Error("IP should not be blocked initially")
	}

	// Block the IP
	storage.Block(ip, 1)

	// Should be blocked now
	if !storage.IsBlocked(ip) {
		t.Error("IP should be blocked after Block call")
	}
}

// TestArrayStorage_UnBlock validates unblocking functionality
func TestArrayStorage_UnBlock(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	ip := net.ParseIP("192.168.1.1")

	// Block the IP
	storage.Block(ip, 60)

	// Unblock the IP
	err := storage.UnBlock(ip)
	if err != nil {
		t.Errorf("UnBlock should not return error: %v", err)
	}

	// Should not be blocked anymore
	if storage.IsBlocked(ip) {
		t.Error("IP should not be blocked after UnBlock call")
	}
}

// TestArrayStorage_ExpiredBlock validates that expired blocks are cleaned up
func TestArrayStorage_ExpiredBlock(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	ip := net.ParseIP("192.168.1.1")

	// Block the IP for a very short duration
	storage.ipList[ip.String()] = time.Now().Add(-1 * time.Second)

	// Should not be blocked (expired)
	if storage.IsBlocked(ip) {
		t.Error("IP should not be blocked when block time has expired")
	}

	// Entry should be cleaned up
	if _, exists := storage.ipList[ip.String()]; exists {
		t.Error("Expired block entry should be removed from map")
	}
}

// TestArrayStorage_Concurrency validates thread-safety
func TestArrayStorage_Concurrency(t *testing.T) {
	storage := &ArrayStorage{
		ipList: make(map[string]time.Time),
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent blocking
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ip := net.ParseIP("192.168.1.1")
			storage.Block(ip, 1)
		}(i)
	}

	// Concurrent checking
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ip := net.ParseIP("192.168.1.1")
			storage.IsBlocked(ip)
		}(i)
	}

	// Concurrent unblocking
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ip := net.ParseIP("192.168.1.1")
			storage.UnBlock(ip)
		}(i)
	}

	wg.Wait()
	// If we get here without a race condition or panic, test passes
}

// TestBlockManager_Block validates BlockManager blocking
func TestBlockManager_Block(t *testing.T) {
	mgr := ArrayBlockManager()

	ip := net.ParseIP("10.0.0.1")
	err := mgr.Block(ip, 1)

	if err != nil {
		t.Errorf("Block should not return error: %v", err)
	}

	if !mgr.IsBlocked(ip) {
		t.Error("IP should be blocked after Block call")
	}
}

// TestBlockManager_IsBlocked validates BlockManager block checking
func TestBlockManager_IsBlocked(t *testing.T) {
	mgr := ArrayBlockManager()

	ip := net.ParseIP("10.0.0.1")

	// Should not be blocked initially
	if mgr.IsBlocked(ip) {
		t.Error("IP should not be blocked initially")
	}

	// Block the IP
	mgr.Block(ip, 1)

	// Should be blocked now
	if !mgr.IsBlocked(ip) {
		t.Error("IP should be blocked after Block call")
	}
}

// TestBlockManager_UnBlock validates BlockManager unblocking
func TestBlockManager_UnBlock(t *testing.T) {
	mgr := ArrayBlockManager()

	ip := net.ParseIP("10.0.0.1")

	// Block the IP
	mgr.Block(ip, 60)

	// Unblock the IP
	err := mgr.UnBlock(ip)
	if err != nil {
		t.Errorf("UnBlock should not return error: %v", err)
	}

	// Should not be blocked anymore
	if mgr.IsBlocked(ip) {
		t.Error("IP should not be blocked after UnBlock call")
	}
}
