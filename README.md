# memlimiter

Library that helps to limit memory consumption of your Go service.

## Working principles
As of Go 1.18, all applications written in Go are leaking memory and will be eventually stopped by OOM killer. The memory leak is because Go runtime knows nothing about the limitations imposed on the process by the operating system (for instance, using cgroups). However, an emergency termination of a process is highly undesirable, as it can lead to data integrity violation, distributed transaction crashes, cache resetting, and even cascading service failure. Therefore, services should degrade gracefully instead of immediate stop due to SIGKILL.

A universal solution for programming languages with automatic memory management comprises two parts:

1. **Garbage collection intensification**. The more often GC starts, the more garbage will be collected, the fewer new physical memory allocations we have to make for the service’s business logic.
2. **Request throttling**. By suppressing some of the incoming requests, we implement the backpressure: the middleware simply cuts off part of the load coming from the client in order to avoid too many memory allocations.

MemLimiter represents a memory budget [automated control system](https://en.wikipedia.org/wiki/Control_system) that helps to keep the memory consumption of a Go service within a predefined limit. 

### Memory budget utilization

The core of the MemLimiter is a special object quite similar to [P-controller](https://en.wikipedia.org/wiki/PID_controller), but with certain specifics (more on that below). Memory budget utilization value acts as an input signal for the controller. We define the utilization as follows:
$$ Utilization = \frac {NextGC} {RSS_{limit} - CGO} $$
where [$NextGC$](https://pkg.go.dev/runtime#MemStats) is a target size for heap, upon reaching which the Go runtime will launch the GC next time.