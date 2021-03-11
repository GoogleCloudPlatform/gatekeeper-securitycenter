// Copyright 2021 Google LLC
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

// Adapted from cloud.google.com/go@v0.57.0/securitycenter/apiv1
//
// Changes:
// - methods pop responses as they are returned
// - remove log import
// - changed github.com/golang/protobuf/proto import to google.golang.org/protobuf/proto
// - renamed clientOpt to clientOptionsForMockServer
// - wrap serv.Serve(lis) in func to avoid errcheck lint error

package securitycenter

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"google.golang.org/api/option"
	securitycenterpb "google.golang.org/genproto/googleapis/cloud/securitycenter/v1"
	iampb "google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type mockSecurityCenterServer struct {
	// Embed for forward compatibility.
	// Tests will keep working if more methods are added
	// in the future.
	securitycenterpb.SecurityCenterServer

	reqs []proto.Message

	// If set, all calls return this error.
	err error

	// responses to return if err == nil
	resps []proto.Message
}

func (s *mockSecurityCenterServer) CreateSource(ctx context.Context, req *securitycenterpb.CreateSourceRequest) (*securitycenterpb.Source, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.Source), nil
}

func (s *mockSecurityCenterServer) CreateFinding(ctx context.Context, req *securitycenterpb.CreateFindingRequest) (*securitycenterpb.Finding, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.Finding), nil
}

func (s *mockSecurityCenterServer) GetIamPolicy(ctx context.Context, req *iampb.GetIamPolicyRequest) (*iampb.Policy, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*iampb.Policy), nil
}

func (s *mockSecurityCenterServer) GetSource(ctx context.Context, req *securitycenterpb.GetSourceRequest) (*securitycenterpb.Source, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.Source), nil
}

func (s *mockSecurityCenterServer) ListFindings(ctx context.Context, req *securitycenterpb.ListFindingsRequest) (*securitycenterpb.ListFindingsResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.ListFindingsResponse), nil
}

func (s *mockSecurityCenterServer) ListSources(ctx context.Context, req *securitycenterpb.ListSourcesRequest) (*securitycenterpb.ListSourcesResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.ListSourcesResponse), nil
}

func (s *mockSecurityCenterServer) SetFindingState(ctx context.Context, req *securitycenterpb.SetFindingStateRequest) (*securitycenterpb.Finding, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*securitycenterpb.Finding), nil
}

func (s *mockSecurityCenterServer) SetIamPolicy(ctx context.Context, req *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	if xg := md["x-goog-api-client"]; len(xg) == 0 || !strings.Contains(xg[0], "gl-go/") {
		return nil, fmt.Errorf("x-goog-api-client = %v, expected gl-go key", xg)
	}
	s.reqs = append(s.reqs, req)
	if s.err != nil {
		return nil, s.err
	}
	var resp proto.Message
	resp, s.resps = s.resps[0], s.resps[1:]
	return resp.(*iampb.Policy), nil
}

// clientOptionsForMockServer is the option tests should use to connect to the test server.
// It is initialized by TestMain.
var clientOptionsForMockServer option.ClientOption

var (
	mockSecurityCenter mockSecurityCenterServer
)

func TestMain(m *testing.M) {
	flag.Parse()

	serv := grpc.NewServer()
	securitycenterpb.RegisterSecurityCenterServer(serv, &mockSecurityCenter)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Could not listen: %+v", err)
		os.Exit(1)
	}
	go func() { _ = serv.Serve(lis) }()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: Could not dial: %+v", err)
		os.Exit(1)
	}
	clientOptionsForMockServer = option.WithGRPCConn(conn)

	os.Exit(m.Run())
}
