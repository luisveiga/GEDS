package main

import (
	"github.com/IBM/gedsmds/internal/config"
	"github.com/IBM/gedsmds/internal/connection/serverconfig"
	"github.com/IBM/gedsmds/internal/logger"
	"github.com/IBM/gedsmds/internal/mdsservice"
	"github.com/IBM/gedsmds/internal/prommetrics"
	"github.com/IBM/gedsmds/protos"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	//"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strconv"
)

func main() {

	
	portNameAux := config.Config.MDSPort
	portNameAux2 := portNameAux[1:len(portNameAux)]
	
	logger.InfoLogger.Println("Metadata Server old port listening is", portNameAux2)
	portNumber, err := strconv.Atoi(portNameAux2)
	logger.InfoLogger.Println("Metadata Server old port number listening is", portNumber)
	if (err == nil) {
		portNumber+=3
		portName :=  ":" + strconv.Itoa(portNumber)
		config.Config.MDSPort = portName
	}

	logger.InfoLogger.Println("Metadata Server new port listening is", config.Config.MDSPort)
	
	metrics := &prommetrics.Metrics{}
	if config.Config.PrometheusEnabled {
		registry := prometheus.NewRegistry()
		registry.MustRegister(collectors.NewGoCollector())
		metrics = prommetrics.InitMetrics(registry)
		//go prometheusServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}))
	}
	lis, err := net.Listen("tcp", config.Config.MDSPort)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	maxMessageSize := 64 * 1024 * 1024
	opts := []grpc.ServerOption{grpc.KeepaliveEnforcementPolicy(serverconfig.KAEP),
		grpc.KeepaliveParams(serverconfig.KASP),
		// this overwrites the 4 MB limits on the messages to be received. New value is 64 MB.
		grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize)}
	grpcServer := grpc.NewServer(opts...)
	serviceInstance := mdsservice.NewService(metrics)
	protos.RegisterMetadataServiceServer(grpcServer, serviceInstance)
	logger.InfoLogger.Println("Metadata Server is listening on port", config.Config.MDSPort)
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
}

func prometheusServer(handler http.Handler) {
	logger.InfoLogger.Println("Prometheus endpoint is listening on port", config.Config.PrometheusPort)
	
	http.Handle("/metrics", handler)
	err := http.ListenAndServe(config.Config.PrometheusPort, nil)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	
}
