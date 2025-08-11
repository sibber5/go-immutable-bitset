// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2025 sibber (GitHub: sibber5)

package bitset

// bitset.Set is an immutable bit set.
type Set interface {
	// Has reports whether the bit for the given bit index is set.
	Has(bitIndex uint32) bool

	// Add returns a new bitset.Set with the bit for the given bit index set.
	// The original bitset.Set is not modified.
	Add(bitIndex uint32) Set

	// Remove returns a new bitset.Set with the bit for the given bit index cleared.
	// The original bitset.Set is not modified.
	Remove(bitIndex uint32) Set
}

// New creates and returns a new empty bitset.Set.
func New() Set {
	return bitSet64(0)
}

// bitset.Builder provides a mutable interface for efficiently constructing a bitset
// by setting the bits before creating the final immutable Set.
//
// WARNING: Using the builder instance after calling Build() is not supported and will cause undefined behavior.
type Builder interface {
	// With returns a new Builder with the bit for the given bit index set.
	With(bitIndex uint32) Builder

	// WithMany returns a new Builder with all the bits for the given bit indices set.
	WithMany(bitIndices ...uint32) Builder

	// Build returns the final immutable Set containing all the bits set on this Builder.
	// Using the builder instance after calling Build() is not supported and will cause undefined behavior.
	Build() Set
}

type bitSetBuilder []uint64

func (b bitSet64) With(bitIndex uint32) Builder {
	s := b.Add(bitIndex)
	if bitIndex < 64 {
		return s.(bitSet64)
	}

	return bitSetBuilder(s.(largeBitSet))
}

func (b bitSet64) WithMany(bitIndices ...uint32) Builder {
	var bld Builder = b
	for _, i := range bitIndices {
		bld = bld.With(i)
	}
	return bld
}

func (b bitSet64) Build() Set {
	return b
}

func (b bitSetBuilder) With(bitIndex uint32) Builder {
	idx := int(bitIndex / 64)
	if idx < len(b) {
		b[idx] |= 1 << (bitIndex % 64)
		return b
	}

	return bitSetBuilder(largeBitSet(b).Add(bitIndex).(largeBitSet))
}

func (b bitSetBuilder) WithMany(bitIndices ...uint32) Builder {
	var bld Builder = b
	for _, i := range bitIndices {
		bld = bld.With(i)
	}
	return bld
}

func (b bitSetBuilder) Build() Set {
	return largeBitSet(b)
}

// NewBuilder creates and returns a new bitset.Builder with an initial bit capacity of at least minCapacity.
// You can set bits beyond this capacity and the builder will expand automatically.
func NewBuilder(minCapacity int) Builder {
	if minCapacity <= 64 {
		return bitSet64(0)
	}

	bits := make([]uint64, (minCapacity+63)/64)
	return bitSetBuilder(bits)
}

// Small (≤64 bits)
type bitSet64 uint64

func (b bitSet64) Has(bitIndex uint32) bool {
	if bitIndex >= 64 {
		return false
	}

	return b&(1<<bitIndex) != 0
}

func (b bitSet64) Add(bitIndex uint32) Set {
	if bitIndex < 64 {
		return b | (1 << bitIndex)
	}

	// Upgrade to largeBitSet
	idx := int(bitIndex / 64)
	newBits := make([]uint64, idx+1)
	newBits[0] = uint64(b)
	newBits[idx] |= 1 << (bitIndex % 64)
	return largeBitSet(newBits)
}

func (b bitSet64) Remove(bitIndex uint32) Set {
	if bitIndex >= 64 {
		return b
	}

	return b &^ (1 << bitIndex)
}

// Large (>64 bits)
type largeBitSet []uint64 // immutable - always copied on modification

func (b largeBitSet) Has(bitIndex uint32) bool {
	idx := int(bitIndex / 64)
	if idx >= len(b) {
		return false
	}

	return b[idx]&(1<<(bitIndex%64)) != 0
}

func (b largeBitSet) Add(bitIndex uint32) Set {
	idx := int(bitIndex / 64)
	newBits := make([]uint64, max(len(b), idx+1))
	copy(newBits, b)
	newBits[idx] |= 1 << (bitIndex % 64)
	return largeBitSet(newBits)
}

func (b largeBitSet) Remove(bitIndex uint32) Set {
	idx := int(bitIndex / 64)
	if idx >= len(b) {
		return b
	}

	lastIdx := len(b) - 1
	for lastIdx >= 0 && (b[lastIdx] == 0 || (lastIdx == idx && b[lastIdx] == 1<<(bitIndex%64))) {
		lastIdx--
	}

	if lastIdx <= 0 {
		b := b[0]
		if b != 0 && idx == 0 {
			b &^= (1 << bitIndex)
		}
		return bitSet64(b)
	}

	bits := b[:(lastIdx + 1)]
	newBits := make([]uint64, len(bits))
	copy(newBits, bits)
	if idx < len(newBits) {
		newBits[idx] &^= 1 << (bitIndex % 64)
	}
	return largeBitSet(newBits)
}
