# Performance Benchmark Tests

This directory contains comprehensive benchmark tests for DBSmith's table rendering performance, focusing on the `QueryResultContent` component and its LRU cell cache.

## Running Benchmarks

```bash
# Run all benchmarks
cd tests/benchmark
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkGetCellSequential -benchmem

# Run with more iterations for accuracy
go test -bench=. -benchmem -benchtime=10s
```

## Benchmark Coverage

### 1. Initialization Performance (`BenchmarkNewQueryResultContent`)
Measures the overhead of creating a `QueryResultContent` instance for different result sizes:
- Small: 100 rows × 10 columns
- Medium: 1,000 rows × 20 columns
- Large: 10,000 rows × 20 columns
- Wide: 1,000 rows × 50 columns

**Key Finding**: Initialization is extremely fast (~30ns) and O(1) regardless of result size, since the cache is lazy-loaded.

### 2. Sequential Cell Access (`BenchmarkGetCellSequential`)
Simulates scrolling through the table by accessing cells sequentially. This represents the common case where users scroll through results.

**Key Finding**: Cache performs excellently with sequential access (~43-50ns per cell for small/medium datasets). Large datasets (10,000 rows) show increased cache misses (~410ns per cell, 4 allocs) as cells get evicted from the 4,000-entry cache.

### 3. Random Cell Access (`BenchmarkGetCellRandom`)
Worst-case scenario where cells are accessed randomly, maximizing cache misses.

**Key Finding**: Performance degrades significantly with random access on larger datasets:
- Small (100 rows): ~50ns, 0 allocs (everything fits in cache)
- Medium (1,000 rows): ~308ns, 3 allocs per access
- Large (10,000 rows): ~503ns, 3 allocs per access

### 4. Cache Hit Performance (`BenchmarkGetCellCacheHits`)
Measures performance when repeatedly accessing a small subset of cells (hot spot).

**Key Finding**: Cache hit rate is excellent at ~43ns per access with 0 allocations, regardless of total dataset size. This validates the LRU cache design for typical scrolling patterns.

### 5. LRU Cache Operations (`BenchmarkLRUCacheOperations`)
Direct measurement of the LRU cache performance with different cache sizes and operation counts:
- Cache 1,000 entries with 100 unique operations
- Cache 4,000 entries with 500 unique operations
- Cache 4,000 entries with 5,000 unique operations (heavy eviction)

**Key Finding**: Cache eviction overhead is minimal (~180-190ns for Set+Get, 2 allocs). Even under heavy eviction pressure, performance remains consistent.

### 6. Alternating Row Colors (`BenchmarkApplyAlternatingRowColors`)
Measures the overhead of applying alternating row background colors.

**Key Finding**: Minimal overhead (~2% difference between with/without alternating colors). The feature is essentially "free" from a performance perspective.

### 7. Cell Value Formatting (`BenchmarkCellValueFormatting`)
Tests the overhead of formatting different value types (short strings, long strings requiring truncation, mixed types).

**Key Finding**: Formatting overhead is negligible (~47ns) and consistent across all value types. String truncation does not add measurable overhead.

### 8. Full Table Render (`BenchmarkFullTableRender`)
Simulates rendering a visible viewport and scrolling through the table:
- 10×10 viewport in 1,000×20 table
- 50×20 viewport in 10,000×50 table

**Key Finding**: Viewport rendering is fast:
- Small viewport: ~41µs to render 100 cells
- Large viewport: ~109µs to render 1,000 cells

This translates to excellent interactive performance even with very large result sets.

## Performance Characteristics Summary

### Strengths
1. **O(1) initialization** - No upfront cost for large result sets
2. **Excellent cache hit performance** - 43ns per access, 0 allocations
3. **Efficient memory usage** - LRU cache prevents unbounded growth
4. **Predictable behavior** - Performance scales linearly with visible viewport size, not total data size

### Bottlenecks
1. **Cache misses on random access** - 300-500ns per miss with 3 allocations
2. **Large dataset eviction** - 10,000+ rows can cause frequent evictions during scrolling
3. **Cache mutex contention** - Not measured but could impact concurrent access (unlikely in TUI)

### Recommendations
1. **Current cache size (4,000 entries) is appropriate** for typical TUI usage
   - Supports viewing ~200 rows × 20 columns with buffer
   - Larger caches would provide diminishing returns

2. **Consider implementing batch cell creation** for initial viewport render
   - Pre-populate cache with visible cells to avoid initial misses
   - Would improve first render time for large result sets

3. **Monitor memory usage** with extremely large result sets
   - QueryResult holds all data in memory
   - Consider streaming or pagination for 100,000+ row results

## Performance Thresholds

Based on these benchmarks, the following thresholds are recommended:

| Result Size | Expected Performance | User Experience |
|-------------|---------------------|-----------------|
| < 1,000 rows | Excellent (<50µs viewport render) | Instantaneous scrolling |
| 1,000 - 10,000 rows | Good (50-100µs viewport render) | Smooth scrolling |
| > 10,000 rows | Acceptable (100-200µs viewport render) | Noticeable but usable |
| > 50,000 rows | Consider pagination | Memory and scroll performance concerns |

The `LargeResultThreshold` constant (1,000 rows) appropriately triggers progress indicators for datasets where performance may be noticeable.
