package remote

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/context"

	pb "internal/remote/api"
)

type server struct {
	mem map[uint64][]byte
}

func (s *server) Alloc(ctx context.Context, in *pb.AllocRequest) (*pb.AllocResponse, error) {
	id := (1 << 63) & rand.Uint64()
	s.mem[id] = make([]byte, in.Size)
	return &pb.AllocResponse{Id: id}, nil
}

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	v, ok := s.mem[in.Id]
	if !ok {
		return nil, fmt.Errorf("no allocation with id %v", in.Id)
	}
	return &pb.GetResponse{Value: v}, nil
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*pb.SetResponse, error) {
	_, ok := s.mem[in.Id]
	if !ok {
		return nil, fmt.Errorf("no allocation with id %v", in.Id)
	}
	s.mem[in.Id] = in.Value
	return &pb.SetResponse{}, nil
}

func newServer() *server {
	s := new(server)
	s.mem = make(map[uint64][]byte)
	return s
}
