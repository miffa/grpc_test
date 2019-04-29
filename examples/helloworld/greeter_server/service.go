package main

import (
	"context"

	//"gitlab.10101111.com/oped/dbms_lib/logrus"
	"github.com/sirupsen/logrus"

	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// SvServer is used to implement helloworld.GreeterSvServer.
type SvServer struct{}

// SayHello implements helloworld.GreeterSvServer
func (s *SvServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	logrus.Debugf("Received: %v", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *SvServer) SayYml(ctx context.Context, in *pb.YmlRequest) (*pb.YmlReply, error) {
	logrus.Debugf("Received: %v", in)
	return &pb.YmlReply{Message: "Hello  hahahahahah"}, nil
}
