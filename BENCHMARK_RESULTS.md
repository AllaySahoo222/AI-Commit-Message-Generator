# Benchmark Results - Before and After

This file contains the raw benchmark output before and after optimizations.

## Before Optimizations

### Git Operations
```
goos: darwin
goarch: arm64
pkg: ai-commit-message-generator/internal/git
cpu: Apple M4 Pro
BenchmarkIsInsideRepo-14                   	   35726	     36601 ns/op	    4320 B/op	      49 allocs/op
BenchmarkHasStagedChanges_NoChanges-14     	    5126	    221593 ns/op	  142076 B/op	     443 allocs/op
BenchmarkHasStagedChanges_OneFile-14       	    3990	    301920 ns/op	  185674 B/op	     631 allocs/op
BenchmarkHasStagedChanges_ManyFiles-14     	     786	   1611685 ns/op	 2032668 B/op	    5364 allocs/op
BenchmarkGetStagedDiff_Small-14            	    2794	    398080 ns/op	  195794 B/op	     777 allocs/op
BenchmarkGetStagedDiff_Medium-14           	    1596	    706029 ns/op	  453712 B/op	    2259 allocs/op
BenchmarkGetStagedDiff_Large-14            	     446	   2617496 ns/op	 2852787 B/op	   23075 allocs/op
BenchmarkGetStagedDiff_VeryLarge-14        	     132	   9076166 ns/op	12284544 B/op	  106640 allocs/op
BenchmarkGetStagedDiff_ModifiedFiles-14    	     950	   1238426 ns/op	  853423 B/op	    6025 allocs/op
PASS
ok  	ai-commit-message-generator/internal/git	14.462s
```

### Config Operations
```
goos: darwin
goarch: arm64
pkg: ai-commit-message-generator/internal/config
cpu: Apple M4 Pro
BenchmarkLoadRules_WithFile-14       	   39555	     32447 ns/op	    2282 B/op	      15 allocs/op
BenchmarkLoadRules_WithoutFile-14    	   68166	     18847 ns/op	    1330 B/op	      11 allocs/op
BenchmarkFindRepoRoot_Shallow-14     	   68608	     17125 ns/op	    1026 B/op	       8 allocs/op
BenchmarkFindRepoRoot_Deep-14        	   51211	     23595 ns/op	    3650 B/op	      28 allocs/op
BenchmarkFindRepoRoot_VeryDeep-14    	   41224	     30488 ns/op	    6066 B/op	      48 allocs/op
PASS
ok  	ai-commit-message-generator/internal/config	8.001s
```

## After Optimizations

### Git Operations
```
goos: darwin
goarch: arm64
pkg: ai-commit-message-generator/internal/git
cpu: Apple M4 Pro
BenchmarkIsInsideRepo-14                   	   87345	     15081 ns/op	     562 B/op	       5 allocs/op
BenchmarkHasStagedChanges_NoChanges-14     	    5965	    194022 ns/op	   70396 B/op	     334 allocs/op
BenchmarkHasStagedChanges_OneFile-14       	    4447	    264486 ns/op	  113141 B/op	     521 allocs/op
BenchmarkHasStagedChanges_ManyFiles-14     	     781	   1459542 ns/op	 1931417 B/op	    5247 allocs/op
BenchmarkGetStagedDiff_Small-14            	    3176	    371003 ns/op	  122199 B/op	     635 allocs/op
BenchmarkGetStagedDiff_Medium-14           	    2050	    566496 ns/op	  349245 B/op	    1075 allocs/op
BenchmarkGetStagedDiff_Large-14            	     632	   1863577 ns/op	 2404362 B/op	    2663 allocs/op
BenchmarkGetStagedDiff_VeryLarge-14        	     193	   6276543 ns/op	 9987573 B/op	    5782 allocs/op
BenchmarkGetStagedDiff_ModifiedFiles-14    	    1174	    919725 ns/op	  679399 B/op	    1725 allocs/op
PASS
ok  	ai-commit-message-generator/internal/git	13.259s
```

### Config Operations
```
goos: darwin
goarch: arm64
pkg: ai-commit-message-generator/internal/config
cpu: Apple M4 Pro
BenchmarkLoadRules_WithFile-14       	   72364	     16762 ns/op	    1026 B/op	       8 allocs/op
BenchmarkLoadRules_WithoutFile-14    	   64639	     18650 ns/op	    1330 B/op	      11 allocs/op
BenchmarkFindRepoRoot_Shallow-14     	   69954	     17136 ns/op	    1026 B/op	       8 allocs/op
BenchmarkFindRepoRoot_Deep-14        	   50989	     23355 ns/op	    3650 B/op	      28 allocs/op
BenchmarkFindRepoRoot_VeryDeep-14    	   42728	     28377 ns/op	    6082 B/op	      48 allocs/op
PASS
ok  	ai-commit-message-generator/internal/config	7.474s
```

## Test Environment

- **OS**: darwin (macOS)
- **Architecture**: arm64
- **CPU**: Apple M4 Pro
- **Go Version**: 1.23.0 (toolchain go1.24.2)
