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
		portNumber+=1
		portName :=  ":" + strconv.Itoa(portNumber)
		config.Config.MDSPort = portName
		logger.InfoLogger.Println("Metadata Server new port to connect is", portName)
	}
	
	ex := mockgedsclient.NewExecutor()

	logger.InfoLogger.Println("---@LIST GATEWAYS")
	ex.ListMDSGateways()

	logger.InfoLogger.Println("---@LIST BUCKETS")
	ex.ListBuckets()
	logger.InfoLogger.Println("---@LIST OBJECT STORE")
	ex.ListObjectStore()
	logger.InfoLogger.Println("---@LIST OBJECTS")
	ex.ListObjects()

	logger.InfoLogger.Println("---@SUBSCRITIONS")
	ex.Subscribe()
	go ex.SubscriberStream()
	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	// go ex.SentUpdateAndCreate()
	// LV END
	
	// LV - exclusive to MCKA
	logger.InfoLogger.Println("---@CREATE BUCKET")
	ex.CreateBucket()
	// LV END 
	logger.InfoLogger.Println("---@LIST BUCKETS")
	ex.ListBuckets()

	logger.InfoLogger.Println("---@LOOKUP BUCKET")
	ex.LookUpBucket()
	logger.InfoLogger.Println("---@LOOKUP BUCKET2")
	ex.LookUpBucket2()

	logger.InfoLogger.Println("---@LOOKUP")
	ex.Lookup()
	// LV - exclusive to MCKA
	logger.InfoLogger.Println("---@CREATE OBJECT")
	ex.CreateObject()
	// LV END
	logger.InfoLogger.Println("---@LIST OBJECTS")
	ex.ListObjects()
	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	logger.InfoLogger.Println("---@UPDATE OBJECT")
	ex.UpdateObject()
	// LV END
	logger.InfoLogger.Println("---@LIST OBJECTS")
	ex.ListObjects()
	logger.InfoLogger.Println("---@LOOKUP")
	ex.Lookup()

	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	// go ex.SentUpdateAndCreate()
	// LV END

	// LV - exclusive to MCKA
	logger.InfoLogger.Println("---@CREATE OBJECT")
	ex.CreateObject()
	
	ex.ListObjects2()

	time.Sleep(30 * time.Second)
	ex.Unsubscribe()

}
