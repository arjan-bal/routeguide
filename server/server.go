// A gRPC server that hosts the routeguide service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"encoding/json"

	pb "github.com/arjan-bal/routeguide"
	"github.com/arjan-bal/routeguide/data"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

var (
	port = flag.Int("port", 50051, "The server port")
	tls  = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
	savedFeatures []pb.Feature
	routeNotes    map[string][]*pb.RouteNote
	mu            sync.Mutex // protects routeNotes
}

func (s *routeGuideServer) GetFeature(ctx context.Context, pt *pb.Point) (*pb.Feature, error) {
	for index := range s.savedFeatures {
        feature := &s.savedFeatures[index]
		if proto.Equal(feature.Location, pt) {
			return proto.Clone(feature).(*pb.Feature), nil
		}
	}
	// No feature was found, return an unnamed feature
	return &pb.Feature{Location: pt}, nil
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{
		routeNotes: make(map[string][]*pb.RouteNote),
	}
	featuresFilePath := data.Path("mock_features.json")
	data, err := os.ReadFile(featuresFilePath)
	if err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
	if err := json.Unmarshal(data, &s.savedFeatures); err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}

	var opts []grpc.ServerOption
	if *tls {
		log.Print("Server is using TLS")
		certFile := data.Path("x509/server_cert.pem")
		keyFile := data.Path("x509/server_key.pem")
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	log.Printf("Starting to listen")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Server terminated due to errorr: %v", err)
	}
}
