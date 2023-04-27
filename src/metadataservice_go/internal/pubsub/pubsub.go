package pubsub

import (
	"github.com/IBM/gedsmds/internal/keyvaluestore"
	"github.com/IBM/gedsmds/internal/logger"
	"github.com/IBM/gedsmds/protos"
	"strings"
	"sync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
	"context"
	"io"
)

func InitService(kvStore *keyvaluestore.Service) *Service {
	service := &Service{
		kvStore: kvStore,

		subscribersStreamLock: &sync.RWMutex{},
		subscriberStreams:     map[string]*SubscriberStream{},

		subscribedItemsLock: &sync.RWMutex{},
		subscribedItems:     map[string][]string{},

		subscribedPrefixLock: &sync.RWMutex{},
		subscribedPrefix:     map[string]map[string][]string{},

		Publication: make(chan *protos.SubscriptionStreamResponse, channelBufferSize),
	}
	go service.runPubSubEventListeners()
	return service
}

func (s *Service) runPubSubEventListeners() {
	for {
		go s.matchPubSub(<-s.Publication)
	}
}

// LV new definitions
var KACP = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             10 * time.Second, // wait 30 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}


func factoryNode(ip string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(KACP)}
	//conn, err := grpc.Dial(ip+config.Config.MDSPort, opts...)
	conn, err := grpc.Dial(ip, opts...)
	if err != nil {
		logger.FatalLogger.Fatalln("Failed to start gRPC connection:", err)
	}
	return conn, err
}

// LV END 

func (s *Service) Subscribe(subscription *protos.SubscriptionEvent) error {
	// LV: subscribe the event locally
	logger.InfoLogger.Println("SUBSCRIBE to preemptively remove and then re-add: ",subscription)
	s.removeSubscriber(subscription)
	s.SubscribeAux(subscription)

	//LV: iterate gateways and asynchronously subscribe the event at it
	s.kvStore.GatewayConfigsLock().RLock()
	defer s.kvStore.GatewayConfigsLock().RUnlock()
	
	logger.InfoLogger.Println("service data being accessed in pubsub", s.kvStore.GatewayConfigs())
	
	for _, gatewayConfig := range s.kvStore.GatewayConfigs() {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		remoteAddress := strings.Clone(gatewayConfig.RemoteAddress)
		conn, err := factoryNode(remoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway in pubsub", remoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway in pubsub", remoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		
		result, err := client.SubscribeAux(context.Background(), subscription)
			if err == nil {
				logger.InfoLogger.Println("pubsub subscription returned from gateway", remoteAddress, result.Code)
				logger.InfoLogger.Println("pubsub subscription returned finally",  result.Code)
			}
		
	}
	return nil
}


// LV new gateway definitions
func (s *Service) SubscribeAux(subscription *protos.SubscriptionEvent) error {
	logger.InfoLogger.Println("SUBSCRIBE-AUX: ", subscription)	
	
	subscribedItemId, err := s.createSubscriptionKey(subscription)
	if err != nil {
		logger.ErrorLogger.Println("SUBSCRIBE-AUX: no subscription key possible", subscription)	
		return err
	}
	if subscription.SubscriptionType == protos.SubscriptionType_BUCKET || subscription.SubscriptionType == protos.SubscriptionType_OBJECT {
		s.subscribedItemsLock.Lock()
		logger.InfoLogger.Println("SUBSCRIBE-AUX: BEFORE", s.subscribedItems, subscription.SubscriberID)	

		s.subscribedItems[subscribedItemId] = append(s.subscribedItems[subscribedItemId], subscription.SubscriberID)
		
		logger.InfoLogger.Println("SUBSCRIBE-AUX: AFTER", s.subscribedItems, subscription.SubscriberID)	
		s.subscribedItemsLock.Unlock()
	} else if subscription.SubscriptionType == protos.SubscriptionType_PREFIX {
		s.subscribedPrefixLock.Lock()
		if _, ok := s.subscribedPrefix[subscription.BucketID]; !ok {
			s.subscribedPrefix[subscription.BucketID] = map[string][]string{}
		}
		s.subscribedPrefix[subscription.BucketID][subscribedItemId] = append(s.subscribedPrefix[subscription.BucketID][subscribedItemId], subscription.SubscriberID)
		s.subscribedPrefixLock.Unlock()
	}
	s.subscribersStreamLock.Lock()
	defer s.subscribersStreamLock.Unlock()
	if streamer, ok := s.subscriberStreams[subscription.SubscriberID]; ok {
		streamer.subscriptionsCounter++
	} else {
		s.subscriberStreams[subscription.SubscriberID] = &SubscriberStream{
			subscriptionsCounter: 1,
		}
	}
	return nil
}

// LV END

func (s *Service) SubscribeStream(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamServer) error {

	// LV: create the stream to the client 
	s.SubscribeStreamAux(subscription, stream)
	
	//LV: iterate gateways and create a stream from each to this MDS
	s.kvStore.GatewayConfigsLock().RLock()
	defer s.kvStore.GatewayConfigsLock().RUnlock()
	
	logger.InfoLogger.Println("service data being accessed in pubsub", s.kvStore.GatewayConfigs())	
	
	for _, gatewayConfig := range s.kvStore.GatewayConfigs() {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		
		remoteAddress := strings.Clone(gatewayConfig.RemoteAddress)
		
		go s.SubscribeStreamAuxClient(subscription, remoteAddress, stream)
	}
return nil	
}

// LV new definitions 
func (s *Service) SubscribeStreamAuxClient(subscription *protos.SubscriptionStreamEvent, remoteAddress string,
	stream protos.MetadataService_SubscribeStreamServer) error {

	conn, err := factoryNode(remoteAddress)
	if conn == nil || err != nil {
		logger.ErrorLogger.Println("Error connection to MDSGategway in pubsub stream aux client", remoteAddress)
	}
	logger.InfoLogger.Println("Setup connection to gateway in pubsub stream aux client", remoteAddress)
	
	client := protos.NewMetadataServiceClient(conn)

	streamer, err := client.SubscribeStreamAux(context.Background(), &protos.SubscriptionStreamEvent{
			SubscriberID: subscription.SubscriberID,
		})
		logger.InfoLogger.Println("subscribing stream aux client at gateway", remoteAddress)
		for {
			object, err := streamer.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.ErrorLogger.Println(err)
				break
			}
			logger.InfoLogger.Println("Object received from gateway as event payload in subscribestreamaux at MDS", object)
			logger.InfoLogger.Println("Forwarding object received from gateway to client", subscription, stream, object)
			// LV TRACE
			logger.TraceLogger.Println("----;", time.Now().UnixMicro(),";PUB;PUB_PUB_AUX;", object.PublicationType, ";", object.Object.Id.Bucket, ";",object.Object.Id.Key, ";",object.Object.Info.Location, ";",object.Object.Info.Size, ";",object.Object.Info.SealedOffset, ";", subscription.SubscriberID)

			s.sendPublication(object, subscription.SubscriberID)
			//s.matchPubSub(object)
		}
		logger.InfoLogger.Println("got disconnected from stream aux at gateway", remoteAddress)
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
	return nil
}

func (s *Service) SubscribeStreamAux(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamServer) error {
	finished := make(chan bool, 2)
	s.subscribersStreamLock.Lock()

	// LV: look up and reuse stream
	if streamer, ok := s.subscriberStreams[subscription.SubscriberID]; !ok {
		s.subscriberStreams[subscription.SubscriberID] = &SubscriberStream{
			stream:   stream,
			finished: finished,
		}
	} else {
		s.removeSubscriberStream(streamer)
		streamer.stream = stream
		streamer.finished = finished
	}
	s.subscribersStreamLock.Unlock()
	cntx := stream.Context()
	for {
		select {
		case <-finished:
			return nil
		case <-cntx.Done():
			return nil
		}
	}
}
// LV END



func (s *Service) matchPubSub(publication *protos.SubscriptionStreamResponse) {
	var subscribers []string
	bucketID, objectID := s.createSubscriptionKeyForMatching(publication.Object)
	s.subscribedItemsLock.RLock()

	logger.InfoLogger.Println("MATCH PUB-SUB publication: ", publication)
	
	if currentSubscribers, ok := s.subscribedItems[bucketID]; ok {
		subscribers = append(subscribers, currentSubscribers...)
		logger.InfoLogger.Println("MATCH PUB-SUB BUCKET sub found: ", subscribers, currentSubscribers)
	}
	if currentSubscribers, ok := s.subscribedItems[objectID]; ok {
		subscribers = append(subscribers, currentSubscribers...)
		logger.InfoLogger.Println("MATCH PUB-SUB OBJECT sub found: ", subscribers, currentSubscribers)
	}
	s.subscribedItemsLock.RUnlock()
	
	s.subscribedPrefixLock.RLock()
	if subscribersInBucket, ok := s.subscribedPrefix[publication.Object.Id.Bucket]; ok {
		for prefix, currentSubscribers := range subscribersInBucket {
			if strings.HasPrefix(objectID, prefix) {				
				subscribers = append(subscribers, currentSubscribers...)
				logger.InfoLogger.Println("MATCH PUB-SUB PREFIX sub found: ", subscribers, currentSubscribers)
			}
		}
	}
	s.subscribedPrefixLock.RUnlock()

	sentSubscriptions := map[string]bool{}
	for _, subscriberID := range subscribers {
		if _, ok := sentSubscriptions[subscriberID]; !ok {
			logger.InfoLogger.Println("PUB-SUB Sending Publication", publication, subscriberID)
			// LV TRACE
			logger.TraceLogger.Println("----;", time.Now().UnixMicro(),";PUB;PUB_PUB;", publication.PublicationType, ";", publication.Object.Id.Bucket, ";",publication.Object.Id.Key, ";",publication.Object.Info.Location, ";",publication.Object.Info.Size, ";",publication.Object.Info.SealedOffset)
			s.sendPublication(publication, subscriberID)
			sentSubscriptions[subscriberID] = true
		}
	}
}

func (s *Service) sendPublication(publication *protos.SubscriptionStreamResponse, subscriberID string) {
	s.subscribersStreamLock.RLock()
	streamer, ok := s.subscriberStreams[subscriberID]
	s.subscribersStreamLock.RUnlock()

	if !ok || streamer.stream == nil {
		logger.ErrorLogger.Println("subscriber stream not found: " + subscriberID)
		return
	}
		if err := streamer.stream.Send(publication); err != nil {
		logger.ErrorLogger.Println("could not send the publication to subscriber " + subscriberID)
		s.removeSubscriberStreamWithLock(streamer)
	}
}

func (s *Service) removeSubscriberStream(streamer *SubscriberStream) {
	if streamer.stream != nil {
		streamer.finished <- true
		streamer.stream = nil
	}
}

func (s *Service) removeSubscriberStreamWithLock(streamer *SubscriberStream) {
	s.subscribersStreamLock.Lock()
	if streamer.stream != nil {
		streamer.finished <- true
		streamer.stream = nil
	}
	s.subscribersStreamLock.Unlock()
}

func (s *Service) removeSubscriber(unsubscription *protos.SubscriptionEvent) error {
	s.subscribersStreamLock.Lock()
	if streamer, ok := s.subscriberStreams[unsubscription.SubscriberID]; ok {
		streamer.subscriptionsCounter--
		if streamer.subscriptionsCounter <= 0 {
			s.removeSubscriberStream(streamer)
		}
	}
	s.subscribersStreamLock.Unlock()

	subscribedItemId, err := s.createSubscriptionKey(unsubscription)
	if err != nil {
		return err
	}
	// possible bottleneck
	s.subscribedItemsLock.Lock()
	if currentSubscribers, ok := s.subscribedItems[subscribedItemId]; ok {
		s.removeElementFromSlice(currentSubscribers, unsubscription.SubscriberID)
	}
	s.subscribedItemsLock.Unlock()
	s.subscribedPrefixLock.Lock()
	if subscribers, ok := s.subscribedPrefix[unsubscription.BucketID]; ok {
		s.removeElementFromSlice(subscribers[subscribedItemId], unsubscription.SubscriberID)
	}
	s.subscribedPrefixLock.Unlock()
	return nil
}


func (s *Service) Unsubscribe(unsubscription *protos.SubscriptionEvent) error {
	// LV: unsubscribe the event locally
	logger.InfoLogger.Println("UN-SUBSCRIBE to preemptively remove and then re-add: ", unsubscription)
	s.UnsubscribeAux(unsubscription)

	//LV: iterate gateways and asynchronously subscribe the event at it
	s.kvStore.GatewayConfigsLock().RLock()
	defer s.kvStore.GatewayConfigsLock().RUnlock()
	
	logger.InfoLogger.Println("service data being accessed in pubsub", s.kvStore.GatewayConfigs())
	
	for _, gatewayConfig := range s.kvStore.GatewayConfigs() {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		remoteAddress := strings.Clone(gatewayConfig.RemoteAddress)
		conn, err := factoryNode(remoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway in pubsub", remoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway in pubsub", remoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		
		result, err := client.UnsubscribeAux(context.Background(), unsubscription)
			if err == nil {
				logger.InfoLogger.Println("pubsub un-subscription returned from gateway", remoteAddress, result.Code)
				logger.InfoLogger.Println("pubsub un-subscription returned finally",  result.Code)
			}
		
	}
	return nil
}

// LV new definitions

func (s *Service) UnsubscribeAux(unsubscription *protos.SubscriptionEvent) error {
	return s.removeSubscriber(unsubscription)
}

// LV END
