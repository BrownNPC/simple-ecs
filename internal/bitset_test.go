package internal_test

import (
	"sync"
	"testing"

	"github.com/BrownNPC/simple-ecs/internal"
	"slices"
)

func TestSetAndIsSet(t *testing.T) {
	bs := &internal.BitSet{}
	
	// Test basic set/get
	bs.Set(3)
	if !bs.IsSet(3) {
		t.Error("Bit 3 should be set")
	}
	if bs.IsSet(2) {
		t.Error("Bit 2 should not be set")
	}

	// Test auto-expansion
	bs.Set(15) // Needs 2 bytes (index 1)
	if !bs.IsSet(15) {
		t.Error("Bit 15 should be set after expansion")
	}
	if len(bs.Data) != 2 {
		t.Error("Data slice should have expanded to 2 bytes")
	}
}

func TestUnset(t *testing.T) {
	bs := &internal.BitSet{}
	bs.Set(5)
	bs.Unset(5)
	if bs.IsSet(5) {
		t.Error("Bit 5 should be unset")
	}

	// Test unset beyond data length
	bs.Unset(100)
	if bs.IsSet(100) {
		t.Error("Bit 100 should not be set")
	}
}

func TestBitwiseOperations(t *testing.T) {
	t.Run("And", func(t *testing.T) {
		bs1 := &internal.BitSet{}
		bs2 := &internal.BitSet{}
		
		bs1.Set(1)
		bs1.Set(2)
		bs2.Set(2)
		bs2.Set(3)

		bs1.And(bs2)
		if !bs1.IsSet(2) || bs1.IsSet(1) || bs1.IsSet(3) {
			t.Error("AND operation failed")
		}
	})

	t.Run("Or", func(t *testing.T) {
		bs1 := &internal.BitSet{}
		bs2 := &internal.BitSet{}
		
		bs1.Set(1)
		bs2.Set(2)
		bs1.Or(bs2)

		if !bs1.IsSet(1) || !bs1.IsSet(2) {
			t.Error("OR operation failed")
		}
	})

	t.Run("AndNot", func(t *testing.T) {
		bs1 := &internal.BitSet{}
		bs2 := &internal.BitSet{}
		
		bs1.Set(1)
		bs1.Set(2)
		bs2.Set(2)
		bs1.AndNot(bs2)

		if !bs1.IsSet(1) || bs1.IsSet(2) {
			t.Error("AND NOT operation failed")
		}
	})
}

func TestActiveIndices(t *testing.T) {
	bs := &internal.BitSet{}
	bs.Set(0)  // First bit
	bs.Set(7)  // Last bit of first byte
	bs.Set(8)  // First bit of second byte
	bs.Set(15) // Last bit of second byte

	expected := []uint32{0, 7, 8, 15}
	indices := internal.ActiveIndices[uint32](bs)

	if !slices.Equal(indices, expected) {
		t.Errorf("Expected %v, got %v", expected, indices)
	}
}

func TestClone(t *testing.T) {
	bs := &internal.BitSet{}
	bs.Set(5)
	clone := bs.Clone()

	// Modify original
	bs.Set(10)

	if clone.IsSet(10) {
		t.Error("Clone should not be affected by original modification")
	}
	if !clone.IsSet(5) {
		t.Error("Clone should retain original bits")
	}
}

func TestConcurrentAccess(t *testing.T) {
	bs := &internal.BitSet{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			bs.Set(uint(n))
			_ = bs.IsSet(uint(n))
			bs.Unset(uint(n))
		}(i)
	}

	wg.Wait()
}

func TestMinint(t *testing.T) {
	testCases := []struct {
		a, b, expected int
	}{
		{5, 3, 3},
		{3, 5, 3},
		{-1, 0, -1},
		{100, 200, 100},
		{0, 0, 0},
	}

	for _, tc := range testCases {
		result := internal.Minint(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("minint(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("ZeroValue", func(t *testing.T) {
		bs := &internal.BitSet{}
		if bs.IsSet(0) {
			t.Error("New bitset should have no bits set")
		}
	})

	t.Run("HighBitPosition", func(t *testing.T) {
		bs := &internal.BitSet{}
		bs.Set(1023) // Test with large position
		if !bs.IsSet(1023) {
			t.Error("High bit position should be set")
		}
	})
}
