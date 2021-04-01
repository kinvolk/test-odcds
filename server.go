package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"time"

	clustercfg "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	clustersvc "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/golang/protobuf/ptypes"
)

type ODCDS struct {
	l *log.Logger
}

func (s *ODCDS) StreamClusters(scs clustersvc.ClusterDiscoveryService_StreamClustersServer) error {
	return errors.New("not implemented")
}

func (s *ODCDS) DeltaClusters(dcs clustersvc.ClusterDiscoveryService_DeltaClustersServer) error {
	req, err := dcs.Recv()
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	s.l.Printf("Got request:\n%s\n", string(j))

	// TODO: Pause rather than replying if there is nothing new.
	// TODO: Need to loop here. Look at go-control-plane.
	// TODO: First request from Envoy should contain a resource name so we just return that
	// resource (cluster). A 2nd request should come in immediately as an ACK. At this point we
	// should probably pause and/or ignore the request. We need to return "real" stuff only when a
	// resource name is specified.
	// TODO: Return a short TTL and ensure "refreshes" are handled.

	// Construct a response.
	resources := []*discovery.Resource{}
	for _, r := range req.ResourceNamesSubscribe {
		if r == "" {
			s.l.Println("Skipping empty resource name")
			continue
		}

		cluster, err := ptypes.MarshalAny(&clustercfg.Cluster{
			Name:           r,
			ConnectTimeout: ptypes.DurationProto(2 * time.Second),
		})
		if err != nil {
			s.l.Printf("Marshalling cluster config: %v", err)
			continue
		}

		resources = append(resources, &discovery.Resource{
			Name:     r,
			Resource: cluster,
			Version:  "v1",
		})
	}

	nonce, err := nonce()
	if err != nil {
		return err
	}

	resp := &discovery.DeltaDiscoveryResponse{
		Resources:         resources,
		Nonce:             nonce,
		TypeUrl:           "type.googleapis.com/envoy.config.cluster.v3.Cluster",
		SystemVersionInfo: "foo",
	}

	j, err = json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	s.l.Printf("Sending response: %v", string(j))

	return dcs.Send(resp)
}

func (s *ODCDS) FetchClusters(context.Context, *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	return nil, errors.New("not implemented")
}

func NewServer(l *log.Logger) *ODCDS {
	return &ODCDS{
		l: l,
	}
}

func nonce() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
