# TaskQ benchmarks

## Description
Here we have 3 catagories of tests:

1. SmallJSONUnmarshal - CPU load task;
2. SleepAndSmallJSONUnmarshalF - IO + CPU (using 50ms sleep for simulating http request behavior); 
3. SleepF - IO (sleep with different time);

And using 3 different approaches to each of tests:
1. Default Taskq mechanism - spawning goroutines for each task;
2. Spawn goroutine with WaitGroup and without taskq for each task;


```
$ go test -bench . -benchmem
goos: linux
goarch: amd64
pkg: github.com/antonmashko/taskq/benchmarks
cpu: AMD Ryzen 9 5900X 12-Core Processor            

BenchmarkTaskq_SleepAndSmallJSONUnmarshalF-24                 	     524	   2113602 ns/op	     784 B/op	      18 allocs/op
BenchmarkSpawningGoroutines_SleepAndSmallJSONUnmarshalF-24    	 1536621	       782.8 ns/op	     815 B/op	      20 allocs/op

BenchmarkTaskq_SmallJSONUnmarshal-24                          	 3745448	       321.3 ns/op	     766 B/op	      18 allocs/op
BenchmarkSpawningGoroutines_SmallJSONUnmarshal-24             	 2198895	       537.2 ns/op	     728 B/op	      19 allocs/op

BenchmarkTaskq_SleepF/1µs-24                                  	 3130252	       362.4 ns/op	     106 B/op	       1 allocs/op
BenchmarkTaskq_SleepF/50µs-24                                 	   27410	     44000 ns/op	     106 B/op	       1 allocs/op
BenchmarkTaskq_SleepF/1ms-24                                  	   26589	     45105 ns/op	     109 B/op	       1 allocs/op
BenchmarkTaskq_SleepF/50ms-24                                 	     522	   2131463 ns/op	      68 B/op	       1 allocs/op
BenchmarkSpawningGoroutines_SleepF/1µs-24                     	 3058214	       403.4 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50µs-24                    	 2920786	       415.8 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/1ms-24                     	 3008365	       410.7 ns/op	     136 B/op	       3 allocs/op
BenchmarkSpawningGoroutines_SleepF/50ms-24                    	 2327683	       477.7 ns/op	     137 B/op	       3 allocs/op

PASS
ok  	github.com/antonmashko/taskq/benchmarks	20.940s
```
