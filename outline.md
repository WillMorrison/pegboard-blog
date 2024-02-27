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



