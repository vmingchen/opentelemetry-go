// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Package main implements a server for Greeter service.

package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/exporters/dynamicconfig"
	pb "go.opentelemetry.io/otel/exporters/dynamicconfig/v1"
)

const (
	port = ":7777"
)

// server is used to implement pb.DynamicConfigServer
type server struct {
	pb.UnimplementedDynamicConfigServer
}

// GetConfig implemented DynamicConfigServer
func (s *server) GetConfig(ctx context.Context, in *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{MetricConfig: dynamicconfig.GetDefaultConfig(1).MetricConfig}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterDynamicConfigServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
