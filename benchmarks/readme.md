# TaskQ benchmarks

## Description
Here we have 3 catagories of tests:

1. SmallJSONUnmarshal - CPU load task;
2. SleepAndSmallJSONUnmarshalF - IO + CPU (using 50ms sleep for simulating http request behavior); 
3. SleepF - IO (sleep with different time);

And using 3 different approaches to each of tests:
1. Default Taskq mechanism - spawning goroutines for each task;
2. TaskQ.Pool - creating a pool of workers and executing task one by one in predefined goroutines (workers);
3. Spawn goroutine with WaitGroup and without taskq for each task;

## Result
As we can see TaskQ.Pool is very affective for CPU tasks. Predefined goroutines creating time savings on scheduler context switching. If we don't have a lot of CPU work and more IO operations than we should spawn goroutine for each task, go scheduler can deal with it without any problems and performance lost. 

```
$ go test -bench . -benchmem
goos: linux
goarch: amd64
pkg: github.com/antonmashko/taskq/benchmarks
cpu: AMD Ryzen 9 5900X 12-Core Processor            

BenchmarkTaskq_SleepAndSmallJSONUnmarshalF-24                 	 1678869	       685.1 ns/op	     937 B/op	      21 allocs/op
BenchmarkSpawningGoroutines_SleepAndSmallJSONUnmarshalF-24    	 1778110	       697.6 ns/op	     809 B/op	      20 allocs/op
BenchmarkTaskqPool_SleepAndSmallJSONUnmarshalF-24             	     524	   2128613 ns/op	     745 B/op	      18 allocs/op

BenchmarkTaskq_SmallJSONUnmarshal-24                          	 2750661	       419.9 ns/op	     851 B/op	      20 allocs/op
BenchmarkSpawningGoroutines_SmallJSONUnmarshal-24             	 2475520	       489.8 ns/op	     728 B/op	      19 allocs/op
BenchmarkTaskqPool_SmallJSONUnmarshal-24                      	 4286218	       277.2 ns/op	     763 B/op	      18 allocs/op

BenchmarkTaskq_SleepF/1µs-24                                  	 3013660	       364.2 ns/op	     248 B/op	       4 allocs/op
BenchmarkTaskq_SleepF/50µs-24                                 	 3212420	       371.6 ns/op	     262 B/op	       4 allocs/op
BenchmarkTaskq_SleepF/1ms-24                                  	 3511340	       341.1 ns/op	     256 B/op	       4 allocs/op
BenchmarkTaskq_SleepF/50ms-24                                 	 3323528	       349.2 ns/op	     252 B/op	       4 allocs/op
BenchmarkSpawningGoroutines_SleepF/1µs-24                     	 3613648	       403.5 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50µs-24                    	 3612339	       406.8 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/1ms-24                     	 3378266	       423.7 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50ms-24                    	 2637591	       454.2 ns/op	     136 B/op	       3 allocs/op
BenchmarkTaskqPool_SleepF/1µs-24                              	 3839192	       349.3 ns/op	     105 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/50µs-24                             	   29520	     44058 ns/op	     100 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/1ms-24                              	   26545	     45095 ns/op	     109 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/50ms-24                             	     517	   2160258 ns/op	      66 B/op	       1 allocs/op

PASS
ok  	github.com/antonmashko/taskq/benchmarks	42.769s

```
