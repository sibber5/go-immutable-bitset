# bitset

A zero dependency, memory efficient, fast, immutable bit set implementation for Go.

## Installation

```bash
go get github.com/sibber5/go-immutable-bitset
```

## Usage

### Basic Operations

```go
// Create a new empty bitset
bs := bitset.New()

// Set bits (returns a new bitset)
bs = bs.Add(42).Add(17)

// Check if a bit is set
if bs.Has(42) {
    fmt.Println("Bit at index 42 is set")
}

// Clear a bit (returns a new bitset)
bs := bs.Remove(42)
```

### Builder Pattern

Use the Builder pattern to efficiently create a new bitset with multiple bits already set:

```go
// Create a builder with expected capacity
builder := bitset.NewBuilder(1000)

// Set multiple bits
bs := builder.
    With(20).
    With(5).
    With(100).
    WithMany(200, 500, 950).
    Build()
```

### Performance Characteristics

The bitset automatically optimizes its internal representation:

- **Small bitsets (≤64 bits)**: Uses a single `uint64` allocated inline
- **Large bitsets (>64 bits)**: Uses a slice of `uint64` with automatic growth and shrinking, optimized for infrequent `Remove`s

`Has` is always O(1). For small bitsets with <=64 bits, `Add` (as long as `butIndex` is <64) and `Remove` are also O(1), otherwise they run in O(n) worse case (where n is len(slice)).

## Thread Safety

Since all bitset operations return new instances rather than modifying existing ones, bitsets are inherently thread-safe for concurrent reads. However, if you need to update a shared bitset reference, you'll need to handle synchronization yourself.

## License

This project is licensed under the BSD 3-Clause "New" or "Revised" License - see the [LICENSE](LICENSE) file for details.
