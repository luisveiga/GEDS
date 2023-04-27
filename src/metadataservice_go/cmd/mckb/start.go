package main

import (
	"github.com/IBM/gedsmds/internal/mockgedsclient"
	"time"
	//_ "google.golang.org/grpc/encoding/gzip"
	"github.com/IBM/gedsmds/internal/config"
	"strconv"
	"github.com/IBM/gedsmds/internal/logger"
)

func main() {
	
	portNameAux := config.Config.MDSPort
	portNameAux2 := portNameAux[1:len(portNameAux)]
	
	logger.InfoLogger.Println("Metadata Server old port to connect is", portNameAux2)
	portNumber, err := strconv.Atoi(portNameAux2)
	logger.InfoLogger.Println("Metadata Server old port number to connect is", portNumber)
	if (err == nil) {
		portNumber+=2
		portName :=  ":" + strconv.Itoa(portNumber)
		config.Config.MDSPort = portName
		logger.InfoLogger.Println("Metadata Server new port to connect is", portName)
	}
		
	ex := mockgedsclient.NewExecutor()

	ex.RegisterMDSGateway("127.0.0.1:50005")

	logger.InfoLogger.Println("---@LIST GATEWAYS")
	ex.ListMDSGateways()

	ex.ListBuckets()
	ex.ListObjectStore()
	ex.ListObjects()

	logger.InfoLogger.Println("---@SUBSCRITIONS")
	ex.Subscribe2()
	go ex.SubscriberStream2()
	
	// LV - exclusive to MCKB
	ex.CreateBucket2()
	// LV END
	logger.InfoLogger.Println("---@LIST BUCKETS")
	ex.ListBuckets()

	logger.InfoLogger.Println("---@LOOKUP BUCKET")
	ex.LookUpBucket()
	logger.InfoLogger.Println("---@LOOKUP BUCKET2")
	ex.LookUpBucket2()

	ex.Lookup()
	// LV - exclusive to MCKA
	// ex.CreateObject()
	// LV END
	ex.ListObjects()
	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	// ex.UpdateObject2()
	// LV END 
	ex.ListObjects()
	
	ex.Lookup()

	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	// go ex.SentUpdateAndCreate2()
	// LV END 
	ex.ListObjects2()

	time.Sleep(30 * time.Second)
	ex.Unsubscribe2()
}
