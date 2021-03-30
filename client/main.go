package main

import (
	"context"
	"log"

	cluster "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial("localhost:9999", opts...)
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}

	client := cluster.NewClusterDiscoveryServiceClient(conn)

	dc, err := client.DeltaClusters(context.Background())
	if err != nil {
		log.Fatalf("DeltaClusters: %v", err)
	}

	req := &discovery.DeltaDiscoveryRequest{
		InitialResourceVersions: map[string]string{"foo": "bar"},
	}
	if err := dc.Send(req); err != nil {
		log.Fatal(err)
	}

	res, err := dc.Recv()
	if err != nil {
		log.Fatalf("Recv: %v", err)
	}

	log.Printf("Response: %v\n", res.String())
}
