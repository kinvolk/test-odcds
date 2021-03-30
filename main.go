package main

import (
	"log"
	"net"
	"os"

	cluster "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	"google.golang.org/grpc"
)

func main() {
	l := log.New(os.Stdout, "", log.Lmsgprefix)
	srv := NewServer(l)
	grpcSrv := grpc.NewServer()
	lis, _ := net.Listen("tcp", ":9999")

	cluster.RegisterClusterDiscoveryServiceServer(grpcSrv, srv)
	log.Printf("Server listening at %s\n", lis.Addr())
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
