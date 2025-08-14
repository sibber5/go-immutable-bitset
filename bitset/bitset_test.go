// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025 sibber (GitHub: sibber5)

package bitset

import (
	"testing"
)

func TestBitSet64(t *testing.T) {
	var bs Set = New()

	if uint64(bs.(bitSet64)) != 0 {
		t.Error("NewBitSet should be empty")
	}

	// Test immutability on Set
	bs2 := bs.Set(5)
	if bs.Test(5) {
		t.Error("Original bitSet64 should not be modified after Set")
	}
	if !bs2.Test(5) {
		t.Error("New bitSet64 should have the set bit")
	}

	// Test immutability on Clear
	bs3 := bs2.Clear(5)
	if !bs2.Test(5) {
		t.Error("Original bitSet64 should not be modified after Clear")
	}
	if bs3.Test(5) {
		t.Error("New bitSet64 should not have the cleared bit")
	}

	// Test setting an existing bit
	bs4 := bs2.Set(5)
	if !bs4.Test(5) {
		t.Error("Setting an existing bit should not clear it")
	}

	// Test removing a non-existent bit
	bs5 := bs2.Clear(10)
	if !bs5.Test(5) {
		t.Error("Removing a non-existent bit should not affect existing bits")
	}

	// Test removing bit >= 64 is a no-op
	bs6 := bs2.Clear(100)
	if !bs6.Test(5) {
		t.Error("Removing a high bit from bitSet64 should be a no-op")
	}
}

func TestBitSetUpgradeAndDowngrade(t *testing.T) {
	var bs Set = New()
	bs = bs.Set(10)

	// Upgrade to largeBitSet
	largeBs := bs.Set(100)

	if _, ok := largeBs.(largeBitSet); !ok {
		t.Fatalf("BitSet should have upgraded to largeBitSet, but got %T", largeBs)
	}

	if !largeBs.Test(10) {
		t.Error("Upgraded set should retain old bits")
	}
	if !largeBs.Test(100) {
		t.Error("Upgraded set should have the new bit")
	}

	// Test immutability during upgrade
	if bs.Test(100) {
		t.Error("Original bitSet64 should not be modified during upgrade")
	}

	// Downgrade back to bitSet64
	downgradedBs := largeBs.Clear(100)
	if _, ok := downgradedBs.(bitSet64); !ok {
		t.Fatalf("Set should have downgraded to bitSet64, but got %T", downgradedBs)
	}
	if !downgradedBs.Test(10) || downgradedBs.Test(100) {
		t.Error("Downgraded set has incorrect bits")
	}

	// Test immutability during downgrade
	if !largeBs.Test(100) {
		t.Error("Original largeBitSet should not be modified during downgrade")
	}
}

func TestLargeBitSet(t *testing.T) {
	// Start with a large set
	var bs Set = New().Set(100)

	// Test immutability on Set
	bs2 := bs.Set(200)
	if bs.Test(200) {
		t.Error("Original largeBitSet should not be modified after Set")
	}
	if !bs2.Test(100) || !bs2.Test(200) {
		t.Error("New largeBitSet should have old and new bits")
	}

	// Test immutability on Clear
	bs3 := bs2.Clear(100)
	if !bs2.Test(100) {
		t.Error("Original largeBitSet should not be modified after Clear")
	}
	if bs3.Test(100) || !bs3.Test(200) {
		t.Error("New largeBitSet should have correct bits after removal")
	}
}

func TestLargeBitSetDowngrade(t *testing.T) {
	// Create a set that will downgrade upon removal
	bs := New().Set(5).Set(70)

	if !bs.Test(5) || !bs.Test(70) {
		t.Fatal("Initial set for downgrade test is incorrect")
	}

	// Clear the high bit, should cause a downgrade
	downgradedBs := bs.Clear(70)

	if _, ok := downgradedBs.(bitSet64); !ok {
		t.Fatalf("Set should have downgraded to bitSet64, but got %T", downgradedBs)
	}

	if !downgradedBs.Test(5) {
		t.Error("Downgraded set is missing its bit")
	}
	if downgradedBs.Test(70) {
		t.Error("Downgraded set should not have the cleared bit")
	}

	// Test downgrade to empty set
	bsEmpty := New().Set(100).Clear(100)
	if _, ok := bsEmpty.(bitSet64); !ok {
		t.Fatalf("Set should have downgraded to bitSet64, but got %T", bsEmpty)
	}
	if bsEmpty.Test(100) {
		t.Error("Set should be empty after removing its only high bit")
	}
	if bsEmpty.(bitSet64) != 0 {
		t.Error("Set should be zero after removing its only bit")
	}
}

func TestBitSetBuilder(t *testing.T) {
	t.Run("Small (bitSet64) builder", func(t *testing.T) {
		b := NewBuilder(10)
		b = b.With(5).With(10)
		bs := b.Build()

		if _, ok := bs.(bitSet64); !ok {
			t.Errorf("Expected bitSet64 from builder, got %T", bs)
		}
		if !bs.Test(5) || !bs.Test(10) {
			t.Error("Built set from small builder has incorrect bits")
		}
	})

	t.Run("Small (bitSet64) builder that upgrades", func(t *testing.T) {
		b := NewBuilder(0)
		b = b.With(10).With(100)

		if _, ok := b.(bitSetBuilder); !ok {
			t.Fatalf("Builder should have transitioned to bitSetBuilderImpl, but is %T", b)
		}

		bs := b.Build()
		if !bs.Test(10) || !bs.Test(100) {
			t.Error("Set built after builder upgrade has incorrect bits")
		}
		if _, ok := bs.(largeBitSet); !ok {
			t.Errorf("Expected largeBitSet from upgraded builder, but got %T", bs)
		}
	})

	t.Run("Large builder from start", func(t *testing.T) {
		b := NewBuilder(200)
		b = b.With(2047).With(0)
		bs := b.Build()

		if !bs.Test(2047) || !bs.Test(0) {
			t.Error("Built set from large builder has incorrect bits")
		}
		if bs.Test(100) {
			t.Error("Built set should not have bits that weren't set")
		}
	})

	t.Run("Growth of bitSetBuilderImpl", func(t *testing.T) {
		b := NewBuilder(100)
		b = b.With(50) // Stays within capacity

		// Trigger growth, which relies on `bitSetBuilderImpl` -> `largeBitSet` -> `bitSetBuilderImpl` convebsion
		b = b.With(200)

		bs := b.Build()
		if !bs.Test(50) || !bs.Test(200) {
			t.Error("Set built after builder growth has incorrect bits")
		}
	})
}

func TestLargeBitSetClearAndShrink(t *testing.T) {
	// Create a set with gaps to test slice trimming
	// bits will be {0, bit for 70, 0, bit for 200}
	bs := New().Set(70).Set(200)

	// Clear 200, which should shrink the backing slice
	bs2 := bs.Clear(200)

	if !bs2.Test(70) {
		t.Error("Set should still have bit 70 after shrinking")
	}
	if bs2.Test(200) {
		t.Error("Set should not have bit 200 after removal")
	}

	// Verify internal slice length (by casting)
	if lbs, ok := bs2.(largeBitSet); ok {
		// Expect length 2 for bits up to 127
		expectedLen := (70 / 64) + 1
		if len(lbs) != expectedLen {
			t.Errorf("Backing slice did not shrink correctly. want len %d, got %d", expectedLen, len(lbs))
		}
	} else {
		t.Errorf("Expected largeBitSet after shrinking, but got %T", bs2)
	}
}
