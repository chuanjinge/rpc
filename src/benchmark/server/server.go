package main

import (
	testpb "benchmark/proto_pb_test"
	"flag"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"os"
	"os/signal"
	"rpc"
	"runtime/pprof"
	"syscall"
)

func ServiceProcessPayload(r *rpc.Router, name string, p rpc.Payload) rpc.Payload {
	if req, ok := p.(*testpb.TestReq); ok {
		rep := testpb.NewTestRep()
		rep.Id = req.Id
		return rep
	} else if b, ok := p.([]byte); ok {

		req := testpb.NewTestReq()
		if proto.Unmarshal(b, req) != nil {
			panic("proto.Unmarshal error")
		}
		rep := testpb.NewTestRep()
		rep.Id = req.Id
		return rep
	} else {
		panic("ServiceProcessPayload receieve wrong info")
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var address = flag.String("address", "127.0.0.1:10000", "benchmark server address")

func main() {
	flag.Parse()

	// profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// protocol
	hf := rpc.NewRPCHeaderFactory(rpc.NewProtobufFactory())

	r, err := rpc.NewRouter(nil, ServiceProcessPayload)
	if err != nil {
		fmt.Println(err)
		return
	}

	r.Run()
	defer r.Stop()

	if err := r.ListenAndServe("benchmark-server", "tcp", *address, hf, nil); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("benchmark server listen on: ", *address)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

forever:
	for {
		select {
		case s := <-sig:
			fmt.Printf("Signal (%d) received, stopping\n", s)
			break forever
		}
	}
}
