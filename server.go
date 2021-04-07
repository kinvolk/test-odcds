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
	corecfg "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointcfg "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
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
	// TODO: Handle concurrent requests.
	for {
		req, err := dcs.Recv()
		if err != nil {
			s.l.Printf("Receiving request: %v", err)
			continue
		}

		j, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			s.l.Printf("Marshaling request JSON: %v", err)
			continue
		}
		s.l.Printf("Got request:\n%s\n", string(j))

		// TODO: Return a short TTL and ensure "refreshes" are handled.

		if req.ResponseNonce != "" {
			// Request is an ACK of a previous response - no need to return a cluster.
			s.l.Printf("Got an ACK with nonce %s", req.ResponseNonce)
			continue
		}

		// Construct a response.
		resources := []*discovery.Resource{}
		for _, r := range req.ResourceNamesSubscribe {
			if r == "" {
				s.l.Println("Skipping empty resource name")
				continue
			}

			cluster, err := ptypes.MarshalAny(makeCluster(r, "127.0.0.1", 8081))
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

		nonce, err := makeNonce()
		if err != nil {
			s.l.Printf("Making nonce: %v", err)
			continue
		}

		resp := &discovery.DeltaDiscoveryResponse{
			Resources:         resources,
			Nonce:             nonce,
			TypeUrl:           "type.googleapis.com/envoy.config.cluster.v3.Cluster",
			SystemVersionInfo: "foo",
		}

		j, err = json.MarshalIndent(resp, "", "  ")
		if err != nil {
			s.l.Printf("Marshaling response JSON: %v", err)
			continue
		}
		s.l.Printf("Sending response:\n%v", string(j))

		err = dcs.Send(resp)
		if err != nil {
			s.l.Printf("Sending response: %v", err)
			continue
		}
	}
}

func (s *ODCDS) FetchClusters(context.Context, *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	return nil, errors.New("not implemented")
}

func NewServer(l *log.Logger) *ODCDS {
	return &ODCDS{
		l: l,
	}
}

func makeNonce() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func makeCluster(name string, host string, port uint32) *clustercfg.Cluster {
	return &clustercfg.Cluster{
		Name:           name,
		ConnectTimeout: ptypes.DurationProto(2 * time.Second),
		LoadAssignment: &endpointcfg.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*endpointcfg.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointcfg.LbEndpoint{
						{
							HostIdentifier: &endpointcfg.LbEndpoint_Endpoint{
								Endpoint: &endpointcfg.Endpoint{
									Address: &corecfg.Address{
										Address: &corecfg.Address_SocketAddress{
											SocketAddress: &corecfg.SocketAddress{
												Protocol: corecfg.SocketAddress_TCP,
												Address:  host,
												PortSpecifier: &corecfg.SocketAddress_PortValue{
													PortValue: port,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
