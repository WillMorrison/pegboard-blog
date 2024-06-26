# Pegboard problem

Place N stones on an NxN grid, such that no pair of stones has the same euclidean distance between them as any other pair of stones.
Solutions exist for N <= 7. We can prove (non-rigorously) that there are no solutions for N >= 15 by observing that the [number of unique distances](https://oeis.org/A160663) asymptotically grows slower than the [number of pairs](https://oeis.org/A000217), and that the number of pairs between 15 stones (120) is greater than the number of unique distances that exist between points on a 15x15 board (119).

The question remains: what about 7 < N < 15?

https://math.stackexchange.com/questions/1208087/maximum-number-of-points-you-can-put-on-grid-n-times-m-with-no-equidistant

## Implementation Optimizations:
- Integer squared distances
- Avoid memory allocation and GC pauses, preallocate
- Parallelization is a wall-time optimization
- Work splitting to account for some threads finishing earlier than others
- Bitstring sets instead of map based

## Algorithm optimizations:
- Pruning with circles and isoceles triangles
  - Pre-calculating these is an implementation optimization
- Fixed ordering to avoid combinatorial explosion
- BFS vs DFS - we only care about existence of solutions, not number. Exploring whole space with BFS guarantees worst case behaviour.
  - BFS also involves a queue, which can add synchronization overhead if shared
- First-stone octant, because we don't care about rotation or reflection.

We want to explore all of these, so there will be multiple versions, interfaces to swap implementations, benchmarking on case 7 and 8. I already know the answers and a bunch of optimization tricks from the prototype, I just want to make it presentable.

## Order of optimization

1. Start with unordered, empty grid, single threaded, map based sets. It's super slow for size=8 (like, more than 12 hours) but finds size 7 solution in a few seconds. We can figure out from this behaviour that searching the whole solution space takes way too long. First because we're searching every permutation of 8 placements (even though we don't care about the order for our problem) - switch to ordered placement and save 8!=40320x time.
2. We can save some more time by eliminating some searching of reflections and rotations by constraining the first stone placed to one octant. 
3. Now we look at a cpu profile. Most of the time is spent doing set operations. Maps are easy to use as sets, but we have constrained set sizes and elements, so we can create a custom bitarray-based set to speed things up even more.
4. Now allocating memory for new objects takes up a large chunk of time (for garbage collection). Preallocating all memory before the search means we don't need to do garbage collection.
5. At this point, checking for separation set membership is taking the most time. We try to avoid doing more work (placement attempts) by keeping a set of places we know stones can't go and skipping placement attempts there.  Even after much work, the pruning implementations have so much overhead for keeping track of the restricted placements that it turns out to be faster to just try and fail.
6. Now we start looking at taking advantage of multiple CPUs. It's reasonably quick to write an implementation that starts multiple searches in parallel, starting each one with a single stone placed. This doesn't give us the expected speedup though, as it turns out that some searches end well before others. By the end of the run, only one CPU is working on the problem while the rest sit idle.
7. By enabling each thread to delegate a section of its work to another, we can keep multiple threads busy right up to the end of 

## Profiling and tracing

There are several flags implemented that let you easily profile and trace.

The `cpuprofile` flag generates a CPU profile, useful for figuring out which lines or functions account for the majority of the procesing time.
```
go build main.go
./main --size=8 -cpuprofile=/tmp/cpu.prof
go tool pprof -http localhost:8888 main /tmp/cpu.prof
```

The `memprofile` flag generates a heap profile, useful for figuring out which lines or functions are allocating memory.
```
go build main.go
./main --size=8 -memprofile=/tmp/mem.prof
go tool pprof -http localhost:8888 -sample_index=alloc_objects main /tmp/mem.prof
```

The `trace` flag provides a way to capture the behaviour of the goroutine scheduler. This is useful when determining if all threads are being used fully.
```
go run . --size=8 -trace=/tmp/trace.out
go tool trace /tmp/trace.out
```