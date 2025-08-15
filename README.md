# A Simple MapReduce Framework in Go

This repository contains a distributed MapReduce framework implemented in Go, providing a foundation for parallel data processing. The system's design allows for user-defined Map and Reduce tasks via a plugin architecture.

## TODO
- [X] Ensure core logic correctness on local
- [ ] Move inter-process communication over the network
- [ ] Run on docker swarm
- [ ] Run on kubernetes
- [ ] Cloud support using EC2 instance(/s) and S3 storage

## Usage

The framework consists of a coordinator and multiple workers. A sequential runner is also available for testing.

### Building Components

The `Makefile` is configured to build all necessary executables and plugins.

```
make all
```


This command generates `mrsequential`, `mrcoordinator`, and `mrworker` binaries in the project root, along with all `.so` plugins in the `mrapps` directory.

### Running a Job

A typical job is executed in three stages:

1.  **Coordinator**: Start the coordinator with input files.

    ```
    ./mrcoordinator ../mrapps/pg-*.txt
    
    ```

2.  **Workers**: Start one or more workers, specifying the plugin.

    ```
    ./mrworker wc.so
    
    ```

3.  **Output**: The final, aggregated output is written to `mr-out-x`.

## Plugin Development

New applications can be developed without recompiling the core worker binary by using Go's plugin (`.so`) build mode.

1.  Create a new directory in `mrapps` (e.g., `mrapps/your_app/`).

2.  Implement your exported `Map` and `Reduce` functions in `main.go`.

    ```
    
    func Map(filename string, contents string) []KeyValue {
        // ...
    }
    
    func Reduce(key string, values []string) string {
        // ...
    }
    
    ```

3.  Add the new plugin target to the `PLUGINS` variable in the `Makefile`.

4.  Build your plugin with `make your_app.so`.

## Testing
See `src/main/test-mr.sh` for detailed testing framework

It uses the `mrsequential.go` for generating the excpected output.
```
./mrsequential wc.so ../mrapps/pg-*.txt
```
