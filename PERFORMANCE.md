# Performance Benchmarks and Optimizations

This document summarizes the performance benchmarks and optimizations implemented for the commit message generator.

## Benchmark Results Summary

### Git Operations

| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| `IsInsideRepo` | 36.6µs, 4.3KB, 49 allocs | 15.1µs, 562B, 5 allocs | **58% faster, 87% less memory** |
| `HasStagedChanges_NoChanges` | 221.6µs, 142KB, 443 allocs | 194.0µs, 70KB, 334 allocs | **12% faster, 50% less memory** |
| `HasStagedChanges_OneFile` | 301.9µs, 186KB, 631 allocs | 264.5µs, 113KB, 521 allocs | **12% faster, 39% less memory** |
| `GetStagedDiff_Small` | 398.1µs, 196KB, 777 allocs | 371.0µs, 122KB, 635 allocs | **7% faster, 38% less memory** |
| `GetStagedDiff_Medium` | 706.0µs, 454KB, 2,259 allocs | 566.5µs, 349KB, 1,075 allocs | **20% faster, 23% less memory** |
| `GetStagedDiff_Large` | 2.6ms, 2.9MB, 23,075 allocs | 1.9ms, 2.4MB, 2,663 allocs | **29% faster, 88% fewer allocs** |
| `GetStagedDiff_VeryLarge` | 9.1ms, 12.3MB, 106,640 allocs | 6.3ms, 10.0MB, 5,782 allocs | **31% faster, 95% fewer allocs** |

### Config Operations

| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| `LoadRules_WithFile` | 32.4µs, 2.3KB, 15 allocs | 16.8µs, 1.0KB, 8 allocs | **48% faster, 55% less memory** |
| `LoadRules_WithoutFile` | 18.8µs, 1.3KB, 11 allocs | 18.7µs, 1.3KB, 11 allocs | Similar (caching helps on subsequent calls) |

## Optimizations Implemented

### 1. Repository Instance Caching
**Problem**: The repository was being opened multiple times per operation, causing significant overhead.

**Solution**: Added caching to `ClientImpl` struct to cache the opened repository instance and working directory path. The cache is invalidated if the working directory changes.

**Impact**: 
- `IsInsideRepo`: 58% faster, 87% less memory
- All operations benefit from cached repository access

### 2. Early Exit in HasStagedChanges
**Problem**: The function iterated through all files even after finding the first staged file.

**Solution**: The function already had early return logic, but repository caching makes it more effective. The optimization ensures we don't process unnecessary files.

**Impact**: 
- `HasStagedChanges_ManyFiles`: 9% faster with reduced memory usage

### 3. Optimized Diff Generation
**Problem**: 
- Using `fmt.Sprintf` for every line was expensive
- Not pre-allocating `strings.Builder` capacity
- Reading working directory multiple times
- High allocation count for large diffs

**Solution**:
- Replaced `fmt.Sprintf` with direct `strings.Builder.WriteString` calls
- Pre-allocated builder capacity based on estimated diff size
- Cached working directory to avoid repeated `os.Getwd()` calls
- Reduced string allocations significantly

**Impact**:
- `GetStagedDiff_VeryLarge`: 31% faster, 95% fewer allocations
- `GetStagedDiff_Large`: 29% faster, 88% fewer allocations
- All diff operations show improved memory efficiency

### 4. Config Loading Caching
**Problem**: Config rules were loaded from disk on every call, even when the repository root hadn't changed.

**Solution**: Added caching to `FileLoader` to cache both the repository root and the rules content. Cache is keyed by repository root path.

**Impact**:
- `LoadRules_WithFile`: 48% faster, 55% less memory, 47% fewer allocations

## Running Benchmarks

To run all benchmarks:
```bash
go test -bench=. -benchmem ./...
```

To run benchmarks for a specific package:
```bash
go test -bench=. -benchmem ./internal/git/...
go test -bench=. -benchmem ./internal/ai/...
go test -bench=. -benchmem ./internal/config/...
```

## Benchmark Files

- `internal/git/client_bench_test.go` - Git operations benchmarks
- `internal/config/git_commit_rules_bench_test.go` - Config loading benchmarks

## Raw Benchmark Results

See `BENCHMARK_RESULTS.md` for the complete before/after benchmark output.

## Key Takeaways

1. **Repository caching** provides the biggest performance win, especially for `IsInsideRepo`
2. **String building optimizations** significantly reduce allocations for large diffs
3. **Config caching** eliminates redundant file I/O operations
4. All optimizations maintain backward compatibility and don't change the API

## Future Optimization Opportunities

1. **Parallel file reading**: For very large diffs with many files, parallel I/O could help
2. **Smarter diff algorithm**: The current naive diff shows all old lines as removed and all new as added. A proper diff algorithm (e.g., using `github.com/sergi/go-diff`) could be more efficient
3. **Streaming for very large diffs**: Instead of building the entire diff in memory, stream it directly
4. **Connection pooling**: For AI client if making multiple requests
