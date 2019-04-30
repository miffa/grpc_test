package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	grpclb "gitlab.10101111.com/oped/DBMS_LIBS/grpclb/ly-grpclb"
	registry "gitlab.10101111.com/oped/DBMS_LIBS/grpclb/ly-grpclb/registry/etcd3"
	etcd3 "go.etcd.io/etcd/clientv3"
)

const (
	address     = "localhost:8080"
	defaultName = "fuck anquan"
)

//  ############################
const (
	BUF_MAX_SIZE = 32 * 1024
	MSG_MAX_SIZE = 4 * 1024 * 1024
)

func main() {
	logrusEntry := logrus.NewEntry(logrus.StandardLogger())
	logopts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	// Make sure that log statements internal to gRPC library are logged using the logrus Logger as well.
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	etcdConfg := etcd3.Config{
		Endpoints: []string{"http://10.104.106.89:2379"},
	}
	r := registry.NewResolver("/grpc-lb-service", "server_", etcdConfg)
	b := grpclb.NewBalancer(r, grpclb.NewKetamaSelector("yml"))
	lbc, err := grpc.Dial("lb-with-etcd",
		grpc.WithInsecure(),
		grpc.WithBalancer(b),
		grpc.WithTimeout(time.Second*5))
	if err != nil {
		log.Printf("grpc dial: %s", err)
		return

	}
	defer lbc.Close()

	lbcr := pb.NewGreeterClient(lbc)
	lbcclicr := healthpb.NewHealthClient(lbc)
	rew, err := lbcr.SayHello(context.WithValue(context.TODO(), "yml", "hahah"),
		&pb.HelloRequest{Name: "Lee leilei"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
		return
	}
	log.Printf("ilb   Greeting: %s", rew.Message)
	resp, err := lbcclicr.Check(context.WithValue(context.TODO(), "yml", "ghghg"), &healthpb.HealthCheckRequest{Service: "laydigaga"})
	if err != nil {
		log.Printf("chek health laydigaga...  err:%v", err)
	} else {
		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			// do something
			log.Printf("server is error")
		} else {
			log.Printf("server is ok:%v", resp)
		}
	}
	//return

	log.Printf("==================tracing from hear=============")

	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)
	tracingopts := []grpc_opentracing.Option{
		//if not ser tracer , use GlobalTracer
		grpc_opentracing.WithTracer(mockTracer),
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{Time: 60 * time.Second, Timeout: 19 * time.Second, PermitWithoutStream: true}),
		grpc.WithMaxMsgSize(MSG_MAX_SIZE),
		grpc.WithWriteBufferSize(BUF_MAX_SIZE),
		grpc.WithReadBufferSize(BUF_MAX_SIZE),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_opentracing.UnaryClientInterceptor(tracingopts...),
			grpc_logrus.UnaryClientInterceptor(logrusEntry, logopts...),
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(grpc_opentracing.StreamClientInterceptor(tracingopts...),
			grpc_logrus.StreamClientInterceptor(logrusEntry, logopts...),
		)),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	cli := healthpb.NewHealthClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//opentracing.StartSpanFromContextWithTracer(ctx, mockTracer, "hello")

	ireq := new(pb.YmlRequest)
	ireq.Bindata = []byte{}
	ireq.Binintdata = []int64{999}
	ireq.Foo = make(map[string]int64)
	ireq.Foo[name] = 99
	ireq.Len = 99
	rer, err := c.SayYml(ctx, ireq)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", rer.Message)

	resp, err = cli.Check(context.TODO(), &healthpb.HealthCheckRequest{Service: "laydigaga..."})
	if err != nil {
		log.Printf("chek health laydigaga...  err:%v", err)
	} else {
		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			// do something
			log.Printf("server is error")
		} else {
			log.Printf("server is ok:%v", resp)
		}
	}
	resp, err = cli.Check(context.TODO(), &healthpb.HealthCheckRequest{Service: "laydigaga"})
	if err != nil {
		log.Printf("chek health err:%v", err)
	} else {
		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			// do something
			log.Printf("server is error")
		} else {
			log.Printf("server is ok:%v", resp)
		}
	}

}
