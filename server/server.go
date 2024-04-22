// A gRPC server that hosts the routeguide service.
package main

import (
 pb "github.com/arjan-bal/routeguide"
)

type routeGuideServer struct {
    pb.UnimplementedRouteGuideServer
}
