package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
	"path/filepath"
)

// Map returns KeyValue
type KeyValue struct {
	Key   string
	Value string
}

type Job struct {
	id, nReduce int
	key,jobType string
}

// We use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	var curr Job
	var err error
	outer:
	for err == nil {
		curr, err = CallExample(curr)
		switch curr.jobType {
		case "map" :
			filename := curr.key
			
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			defer file.Close()

			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}

			var encoders []*json.Encoder
			for i := range curr.nReduce {
				file, err = os.Create(fmt.Sprintf("inter-%v-%v", curr.id, i))
				if err != nil {
					panic("error creating intermediate file")
				}
				defer file.Close()
				encoders = append(encoders, json.NewEncoder(file))
			}

			kva := mapf(filename, string(content))
			sort.Slice(kva, func(i, j int) bool{return kva[i].Key < kva[j].Key})

			for _, kv := range kva {
				hashvalue := ihash(kv.Key) % curr.nReduce
				err := encoders[hashvalue].Encode(&kv)
				if err != nil {
					panic("error writing to intermediate file")
				}
			}
		case "reduce" :
			pattern := "inter-*-" + fmt.Sprint(curr.id)
			filenames, err := filepath.Glob(pattern)
			if err != nil {
				fmt.Println("error getting intermediate files in reduce")
			}

			intermediates := make(map[string][]string)
			for _, filename := range filenames{
				file, err := os.Open(filename)
				if err != nil {
					panic("Error opening intermediate file in reduce")
				}
				defer file.Close()
				decoder :=  json.NewDecoder(file)
				var kv KeyValue
				for {
					if err := decoder.Decode(&kv); err != nil {
						break
					}
					intermediates[kv.Key] = append(intermediates[kv.Key], kv.Value)
				}
			}

			ofilename := "mr-out-" + fmt.Sprint(curr.id)
			ofile, err := os.Create(ofilename)
			if err != nil {
				panic("Output file creation failure in reduce")
			}
			for key := range intermediates {
				output := reducef(key, intermediates[key])
				fmt.Fprintf(ofile, "%v %v\n", key, output)
			}
		case "end" :
			break outer
		default :
			time.Sleep(1 * time.Second)
		}
	}
}

func CallExample(finished Job) (Job, error) {

	args  := Args{Id: finished.id, Key: finished.key, JobType: finished.jobType}
	reply := Reply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Schedule", &args, &reply)
	if !ok {
		fmt.Printf("call failed!\n")
		return Job{}, fmt.Errorf("call failed")
	}
	return Job{id: reply.Id, key:reply.Key, jobType:reply.JobType, nReduce: reply.NReduce}, nil
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
