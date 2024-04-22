// Packange main implements a gRPC client that interacts with a server running
// locally which hosts the route guide server.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "github.com/arjan-bal/routeguide"
	"github.com/arjan-bal/routeguide/data"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	port               = flag.Int("port", 50051, "The server port")
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
	serverAddr         = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

// Quieries the server and prints the feature at the given point.
func printFeature(pt *pb.Point, client pb.RouteGuideClient) {
	log.Printf("Getting feature for point (%d, %d)", pt.Latitute, pt.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.GetFeature(ctx, pt)
	if err != nil {
		log.Fatalf("Failed to get point from server: %v", err)
	}
	log.Println(prototext.Format(feature))
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		caFile := data.Path("x509/ca_cert.pem")
		creds, err := credentials.NewClientTLSFromFile(caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	defer conn.Close()
	client := pb.NewRouteGuideClient(conn)
	// Valid feature.
	printFeature(&pb.Point{Latitute: 409146138, Longitude: -746188906}, client)
	// Invalid feature.
	printFeature(&pb.Point{Latitute: 0, Longitude: 0}, client)
}