package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)


type Task struct{
	key string
	startTime time.Time
	ongoing bool
	finished bool
}

type Coordinator struct {
	nReduce int
	mapToDo, reduceToDo []Task
	mapDone, reduceDone bool
	sync.Mutex
}

func (c *Coordinator) Schedule(args *Args, reply *Reply) error {
	
	if args.JobType == "map" {
		c.mapToDo[args.Id].finished = true
		c.mapToDo[args.Id].ongoing = false
	} else if args.JobType == "reduce" {
		c.reduceToDo[args.Id].finished = true
		c.reduceToDo[args.Id].ongoing = false
	}

	c.Lock()
	defer c.Unlock()
	if !c.mapDone {
		mapdone := true
		for i := range c.mapToDo {
			var timeRunning time.Duration
			if c.mapToDo[i].ongoing {
				timeRunning = time.Since(c.mapToDo[i].startTime)
			}
			if (!c.mapToDo[i].finished && !c.mapToDo[i].ongoing) || timeRunning > 10 * time.Second{
				reply.Id = i
				reply.Key = c.mapToDo[i].key
				reply.JobType = "map"
				reply.NReduce = c.nReduce
				c.mapToDo[i].startTime = time.Now()
				c.mapToDo[i].ongoing = true
				return nil
			}
			mapdone = mapdone && c.mapToDo[i].finished
		}
		if !mapdone {
			*reply = Reply{}
			return nil
		}
		c.mapDone = true
	}

	if !c.reduceDone {
		redcuedone := true
		for i := range c.reduceToDo {
			var timeRunning time.Duration
			if c.reduceToDo[i].ongoing {
				timeRunning = time.Since(c.reduceToDo[i].startTime)
			}
			if (!c.reduceToDo[i].finished && !c.reduceToDo[i].ongoing) || timeRunning > 10 * time.Second{
				reply.Id = i
				reply.JobType = "reduce"
				reply.NReduce = c.nReduce
				c.reduceToDo[i].startTime = time.Now()
				c.reduceToDo[i].ongoing = true
				return nil
			}
			redcuedone = redcuedone && c.reduceToDo[i].finished
		}
		if !redcuedone {
			*reply = Reply{}
			return nil
		}
		c.reduceDone = true
	}

	*reply = Reply{JobType: "end"}

	return nil
}

// starts a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	return c.mapDone && c.reduceDone
}

// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.nReduce = nReduce

	for _, file := range files {
		c.mapToDo = append(c.mapToDo, Task{
				key: file,
				startTime: time.Now(),
				finished: false,
		})
	}
	c.reduceToDo = make([]Task, c.nReduce)

	c.server()
	return &c
}
