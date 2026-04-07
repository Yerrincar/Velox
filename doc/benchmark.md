### Concurrency experiment

In this test we are going to compare the diferent methods of Concurrency that go has to offer and how to optimize
them as much as possible.

We are going to split tests in two:
 - Mock Test: Same amount of files and size than the real files, but in a local folder to compare each Concurrency
 strategy.
 - Real Test: The actual files I want to transfer from the phone to my pc and then to the VM on my home server.

1. First attempt, no Concurrency implemented:

For this implementation, no concurrency was added yet. The program run against the real data and the time of execution
was 3m 30s. 

2. Second attempt, only go routines, no boundaries:

As expected, we faced one of the problems that we may have if we don't control the boundaries of our program. 
Around 500-550 go routines were spawned at the same time, one for each gio copy, which shouldn't be a problem 
per se, but the hundreds of OS processes + heavy I/O contention created are. For this reason, the pc freeze 
when the program is executed.


### Times 

- First version without concurrency: 3m 30s
- First implementation with sempahores and maxConcurrency of 3: 2m 48s
