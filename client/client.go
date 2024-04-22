// Packange main implements a gRPC client that interacts with a server running
// locally which hosts the route guide server.
package main

import (
	"context"
	"flag"
	"io"
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
	log.Printf("Getting feature for point (%d, %d)", pt.Latitude, pt.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feature, err := client.GetFeature(ctx, pt)
	if err != nil {
		log.Fatalf("Failed to get point from server: %v", err)
	}
	log.Println(prototext.Format(feature))
}

// printFeatures lists all the features within the given bounding Rectangle.
func printFeatures(client pb.RouteGuideClient, rect *pb.Rectangle) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("Failed to stream featrues from the server: %v", err)
	}

	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			log.Print("End of stream from the server!")
			break
		}
		if err != nil {
			log.Fatalf("Failed to stream features from server: %v", err)
		}
		log.Printf("Got feature from the server %s\n", prototext.Format(feature))
	}
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
	printFeature(&pb.Point{Latitude: 409146138, Longitude: -746188906}, client)
	// Invalid feature.
	printFeature(&pb.Point{Latitude: 0, Longitude: 0}, client)

	// Looking for features between 40, -75 and 42, -73.
	printFeatures(client, &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	})
}
