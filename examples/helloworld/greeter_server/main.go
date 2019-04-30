package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	//"gitlab.10101111.com/oped/dbms_lib/logrus"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"

	registry "gitlab.10101111.com/oped/dbms_lib/grpclb/ly-grpclb/registry/etcd3"
	etcd3 "go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

var (
	nodeID    = flag.String("nodeID", "nodeX", "node ID, must be unique")
	serv      = flag.String("service", "server_", "server name")
	port      = flag.Int("port", 8080, "listening port")
	ip        = flag.String("ip", "127.0.0.1", "binding ip")
	grpclbDir = flag.String("lb", "/grpc-lb-service", "grpc discovery dir")
)

func main() {
	flag.Parse()
	ipaddr := *ip + ":" + strconv.Itoa(*port)

	uid, _ := uuid.NewV4()
	Serverid := *serv + "_" + *nodeID + "_" + uid.String()
	fmt.Println(Serverid)

	etcdConfg := etcd3.Config{
		Endpoints: []string{"http://10.104.106.89:2379"},
	}

	registrySrv, err := registry.NewRegistry(
		registry.Option{
			EtcdConfig:  etcdConfg,
			RegistryDir: *grpclbDir,
			ServiceName: *serv,
			NodeID:      *nodeID,
			NData: registry.NodeData{
				Addr: ipaddr,
				//Metadata: map[string]string{"weight": "1"},
			},
			Ttl: 2019 * time.Second,
		},
	)

	if err != nil {
		log.Panic(err)
		return

	}

	lis, err := net.Listen("tcp", ipaddr)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	logrusEntry := logrus.NewEntry(logrus.StandardLogger())
	logopts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}
	logrus.SetLevel(logrus.DebugLevel)

	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)
	tracingopts := []grpc_opentracing.Option{
		grpc_opentracing.WithTracer(mockTracer),
	}

	rcustomfunc = RecoveryFromPanic
	rcustomfuncCtx = RecoveryContexyFromPanic
	ropts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(rcustomfunc),
		grpc_recovery.WithRecoveryHandlerContext(rcustomfuncCtx),
	}

	// Make sure that log statements internal to gRPC library are logged using the logrus Logger as well.
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	s := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: 300 * time.Second,
	}),
		grpc.ConnectionTimeout(10*time.Second),
		grpc_middleware.WithUnaryServerChain(
			grpc_opentracing.UnaryServerInterceptor(tracingopts...),
			grpc_recovery.UnaryServerInterceptor(ropts...),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, logopts...),
		),

		grpc_middleware.WithStreamServerChain(
			grpc_opentracing.StreamServerInterceptor(tracingopts...),
			grpc_recovery.StreamServerInterceptor(ropts...),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logopts...),
		),
	)

	pb.RegisterGreeterServer(s, &SvServer{})

	hsrv := health.NewServer()
	hsrv.SetServingStatus("laydigaga", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hsrv)

	go func() {
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("failed to serve: %v", err)
		}
	}()

	registrySrv.Register()
	defer registrySrv.Deregister()

	cg := make(chan os.Signal, 1)
	signal.Notify(cg,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGSTOP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)

	for {
		gs := <-cg
		fmt.Printf("signal coming:%v", gs)
		switch gs {
		case syscall.SIGQUIT,
			syscall.SIGTERM,
			syscall.SIGSTOP,
			syscall.SIGINT,
			syscall.SIGHUP:
			hsrv.Shutdown()
			//hsrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
			s.Stop()

			time.Sleep(3 * time.Second)
			return
		case syscall.SIGUSR2:
			hsrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
			s.Stop()
			hsrv.Shutdown()
			time.Sleep(3 * time.Second)
			return
		case syscall.SIGUSR1:
			hsrv.Shutdown()
			//	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
			s.Stop()
			time.Sleep(3 * time.Second)
			return
		default:
			fmt.Printf("signal coming xxxxx:%v", gs)
		}

	}
}
