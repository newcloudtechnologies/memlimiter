# MemLimiter

Library that helps to limit memory consumption of your Go service.

## Working principles
As of Go 1.18, all applications written in Go are leaking memory and will be eventually stopped by OOM killer. The memory leak is because Go runtime knows nothing about the limitations imposed on the process by the operating system (for instance, using cgroups). However, an emergency termination of a process is highly undesirable, as it can lead to data integrity violation, distributed transaction crashes, cache resetting, and even cascading service failure. Therefore, services should degrade gracefully instead of immediate stop due to SIGKILL.

A universal solution for programming languages with automatic memory management comprises two parts:

1. **Garbage collection intensification**. The more often GC starts, the more garbage will be collected, the fewer new physical memory allocations we have to make for the service’s business logic.
2. **Request throttling**. By suppressing some of the incoming requests, we implement the backpressure: the middleware simply cuts off part of the load coming from the client in order to avoid too many memory allocations.

MemLimiter represents a memory budget [automated control system](https://en.wikipedia.org/wiki/Control_system) that helps to keep the memory consumption of a Go service within a predefined limit. 

### Memory budget utilization

The core of the MemLimiter is a special object quite similar to [P-controller](https://en.wikipedia.org/wiki/PID_controller), but with certain specifics (more on that below). Memory budget utilization value acts as an input signal for the controller. We define the $Utilization$ as follows:
$$ Utilization = \frac {NextGC} {RSS_{limit} - CGO} $$
where:
* $NextGC$ ([from here](https://pkg.go.dev/runtime#MemStats)) is a target size for heap, upon reaching which the Go runtime will launch the GC next time;
* $RSS_{limit}$ is a hard limit for service's physical memory (`RSS`) consumption (so that exceeding this limit will highly likely result in OOM);
* $CGO$ is a total size of heap allocations made beyond `Cgo` borders (within `C`/`C++`/.... libraries).

A few notes about $CGO$ component. Allocations made outside of the Go allocator, of course, are not controlled by the Go runtime in any way. At the same time, the memory consumption limit is common for both Go and non-Go allocators. Therefore, if non-Go allocations grow, all we can do is shrink the memory budget for Go allocations (which is why we subtract $CGO$ from the denominator of the previous expression). If your service uses `Cgo`, you need to figure out how much memory is allocated “on the other side”– otherwise MemLimiter won’t be able to save your service from OOM.

If the service doesn't use `Cgo`, the $Utilization$ formula is simplified to:
$$Utilization = \frac {NextGC} {RSS_{limit}}$$

### Control function

The controller converts the input signal into the control signal according to the following formula:

$$  K_{p} = C \cdot \frac {1} {1 - Utilization} $$

This is not an ordinary definition for a proportional component of the PID-controller, but still the direct proportionality is preserved: the closer the $Utilization$ is to 1 (or 100%), the higher the control signal value. The main purpose of the controller is to prevent a situation in which the next GC launch will be scheduled when the memory consumption exceeds the hard limit (and this will cause OOM).

You can adjust the proportional component control signal strength using a coefficient $C$. In addition, there is optional [exponential averaging](https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average) of the control signal. This helps to smooth out high-frequency fluctuations of the control signal (but it hardly eliminates [self-oscillations](https://en.wikipedia.org/wiki/Self-oscillation)).

The control signal is always saturated to prevent extremal values:

$$ Output = \begin{cases}
\displaystyle 100 \ \ \ K_{p} \gt 100 \\
\displaystyle 0 \ \ \ \ \ \ \ K_{p} \lt 100 \\
\displaystyle K_{p} \ \ \ \ otherwise \\
\end{cases}$$

Finally we convert the dimensionless quantity $Output$ into specific $GOGC$ (for the further use in [`debug.SetGCPercent`](https://pkg.go.dev/runtime/debug#SetGCPercent)) and $Throttling$ (percentage of suppressed requests) values, however, only if the $Utilization$ exceeds the specified limits:


$$ GC = \begin{cases}
\displaystyle Output \ \ \ Utilization \gt DangeZoneGC \\
\displaystyle 100 \ \ \ \ \ \ \ \ \ \ otherwise \\
\end{cases}$$

$$ Throttling = \begin{cases}
\displaystyle Output \ \ \ Utilization \gt DangeZoneThrottling \\
\displaystyle 0 \ \ \ \ \ \ \ \ \ \ \ \ \ \ otherwise \\
\end{cases}$$