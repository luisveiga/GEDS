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

	for {
	logger.InfoLogger.Println("---@CREATE BUCKET")
	ex.CreateBucket()
	time.Sleep(100 * time.Millisecond)
	// LV END 
	logger.InfoLogger.Println("---@CREATE OBJECT")
	ex.CreateObject()
	time.Sleep(100 * time.Millisecond)
	// LV END
	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	logger.InfoLogger.Println("---@UPDATE OBJECT")
	ex.UpdateObject()
	time.Sleep(100 * time.Millisecond)
	// LV END
	// LV - exclusive to MCKA - updating objects creates (new) entry in own MDS
	logger.InfoLogger.Println("---@UPDATE AND CREATE OBJECT")
	ex.SentUpdateAndCreate()
	time.Sleep(100 * time.Millisecond)
	// LV END
	}

}
