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

BenchmarkTaskq_SleepAndSmallJSONUnmarshalF-24                 	  374064	      2692 ns/op	    1120 B/op	      24 allocs/op
BenchmarkSpawningGoroutines_SleepAndSmallJSONUnmarshalF-24    	 1505056	       745.2 ns/op	     809 B/op	      20 allocs/op
BenchmarkTaskqPool_SleepAndSmallJSONUnmarshalF-24             	     516	   2172724 ns/op	     748 B/op	      18 allocs/op

BenchmarkTaskq_SmallJSONUnmarshal-24                          	  740080	      2239 ns/op	     972 B/op	      23 allocs/op
BenchmarkSpawningGoroutines_SmallJSONUnmarshal-24             	 2423046	       499.8 ns/op	     728 B/op	      19 allocs/op
BenchmarkTaskqPool_SmallJSONUnmarshal-24                      	 4214408	       280.2 ns/op	     765 B/op	      18 allocs/op

BenchmarkTaskq_SleepF/1µs-24                                  	  948886	      1816 ns/op	     376 B/op	       7 allocs/op
BenchmarkTaskq_SleepF/50µs-24                                 	  825915	      1966 ns/op	     370 B/op	       7 allocs/op
BenchmarkTaskq_SleepF/1ms-24                                  	  696423	      2160 ns/op	     386 B/op	       7 allocs/op
BenchmarkTaskq_SleepF/50ms-24                                 	  460440	      2424 ns/op	     398 B/op	       7 allocs/op
BenchmarkSpawningGoroutines_SleepF/1µs-24                     	 3290787	       381.0 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50µs-24                    	 3269288	       369.0 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/1ms-24                     	 3282033	       432.1 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50ms-24                    	 2066744	       540.6 ns/op	     137 B/op	       3 allocs/op
BenchmarkTaskqPool_SleepF/1µs-24                              	 3417200	       305.8 ns/op	     109 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/50µs-24                             	   28244	     43166 ns/op	     103 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/1ms-24                              	   26728	     45068 ns/op	     108 B/op	       1 allocs/op
BenchmarkTaskqPool_SleepF/50ms-24                             	     522	   2128463 ns/op	      80 B/op	       1 allocs/op

PASS
ok  	github.com/antonmashko/taskq/benchmarks	32.501s
```
