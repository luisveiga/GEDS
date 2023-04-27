package mdsprocessor

import (
	"context"
	"errors"
	"github.com/IBM/gedsmds/internal/config"
	"github.com/IBM/gedsmds/internal/keyvaluestore"
	"github.com/IBM/gedsmds/internal/logger"
	"github.com/IBM/gedsmds/internal/pubsub"
	"github.com/IBM/gedsmds/protos"
	"google.golang.org/grpc/peer"
	"strings"
)

func InitService() *Service {
	kvStore := keyvaluestore.InitKeyValueStoreService()
	return &Service{
		pubsub:  pubsub.InitService(kvStore),
		kvStore: kvStore,
	}
}

func (s *Service) GetClientConnectionInformation(ctx context.Context) (string, error) {
	if peerInfo, ok := peer.FromContext(ctx); !ok {
		return "", errors.New("client IP could not be parsed")
	} else {
		return strings.Split(peerInfo.Addr.String(), ":")[0], nil
	}
}

func (s *Service) RegisterObjectStore(objectStore *protos.ObjectStoreConfig) error {
	if err := s.kvStore.RegisterObjectStore(objectStore); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListObjectStores() (*protos.AvailableObjectStoreConfigs, error) {
	return s.kvStore.ListObjectStores()
}

func (s *Service) CreateBucket(bucket *protos.Bucket) error {
	if err := s.kvStore.CreateBucket(bucket); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteBucket(bucket *protos.Bucket) error {
	if err := s.kvStore.DeleteBucket(bucket); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListBuckets() (*protos.BucketListResponse, error) {
	return s.kvStore.ListBuckets()
}

func (s *Service) LookupBucket(bucket *protos.Bucket) error {
	if err := s.kvStore.LookupBucket(bucket); err != nil {
		return err
	}
	return nil
}

// LV new gateway functions

func (s *Service) LookupBucketAux(bucket *protos.Bucket) error {
	if err := s.kvStore.LookupBucketAux(bucket); err != nil {
		return err
	}
	return nil
}
// LV END

func (s *Service) CreateObject(object *protos.Object) error {
	if err := s.kvStore.CreateObject(object); err != nil {
		return err
	}
	if config.Config.PubSubEnabled {
		logger.InfoLogger.Println("PUBSUB AT MDS SERVICE - CREATE: ", object, protos.PublicationType_CREATE_OBJECT)
		s.pubsub.Publication <- &protos.SubscriptionStreamResponse{
			Object:          object,
			PublicationType: protos.PublicationType_CREATE_OBJECT,
		}
	}
	return nil
}

func (s *Service) CreateOrUpdateObjectStream(object *protos.Object) {
	if err := s.kvStore.CreateObject(object); err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	if config.Config.PubSubEnabled {
		logger.InfoLogger.Println("PUBSUB AT MDS SERVICE - CREATE_UPDATE: ", object, protos.PublicationType_CREATE_UPDATE_OBJECT)
		s.pubsub.Publication <- &protos.SubscriptionStreamResponse{
			Object:          object,
			PublicationType: protos.PublicationType_CREATE_UPDATE_OBJECT,
		}
	}
}

func (s *Service) UpdateObject(object *protos.Object) error {
	if err := s.kvStore.UpdateObject(object); err != nil {
		return err
	}
	if config.Config.PubSubEnabled {
		logger.InfoLogger.Println("PUBSUB AT MDS SERVICE - UPDATE: ", object, protos.PublicationType_UPDATE_OBJECT)
		s.pubsub.Publication <- &protos.SubscriptionStreamResponse{
			Object:          object,
			PublicationType: protos.PublicationType_UPDATE_OBJECT,
		}
	}
	return nil
}

func (s *Service) DeleteObject(objectID *protos.ObjectID) error {
	if err := s.kvStore.DeleteObject(objectID); err != nil {
		return err
	}
	if config.Config.PubSubEnabled {
		logger.InfoLogger.Println("PUBSUB AT MDS SERVICE - DELETE: ", objectID, protos.PublicationType_DELETE_OBJECT)
		s.pubsub.Publication <- &protos.SubscriptionStreamResponse{
			Object:          &protos.Object{Id: objectID},
			PublicationType: protos.PublicationType_DELETE_OBJECT,
		}
	}
	return nil
}

func (s *Service) DeletePrefix(objectID *protos.ObjectID) error {
	objects, err := s.kvStore.DeleteObjectPrefix(objectID)
	if err != nil {
		return err
	}
	if config.Config.PubSubEnabled {
		for _, object := range objects {
			logger.InfoLogger.Println("PUBSUB AT MDS SERVICE - DELETE_PREFIX: ", object, protos.PublicationType_DELETE_OBJECT)
			s.pubsub.Publication <- &protos.SubscriptionStreamResponse{
				Object:          object,
				PublicationType: protos.PublicationType_DELETE_OBJECT,
			}
		}
	}
	return nil
}

func (s *Service) LookupObject(objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	object, err := s.kvStore.LookupObject(objectID)
	if err != nil {
		return &protos.ObjectResponse{
			Error: &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND},
		}, nil
	}
	return object, nil
}

// LV new gateway functions

func (s *Service) LookupObjectAux(objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	object, err := s.kvStore.LookupObjectAux(objectID)
	if err != nil {
		return &protos.ObjectResponse{
			Error: &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND},
		}, nil
	}
	return object, nil
}

// LV END

func (s *Service) List(objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	return s.kvStore.ListObjects(objectListRequest)
}

// LV new gateway functions
func (s *Service) ListBucketsAux() (*protos.BucketListResponse, error) {
	return s.kvStore.ListBucketsAux()
}


func (s *Service) ListAux(objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	return s.kvStore.ListObjectsAux(objectListRequest)
}


// LV end 

func (s *Service) Subscribe(subscription *protos.SubscriptionEvent) error {
	return s.pubsub.Subscribe(subscription)
}

func (s *Service) SubscribeStream(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamServer) error {
	return s.pubsub.SubscribeStream(subscription, stream)
}

func (s *Service) Unsubscribe(unsubscribe *protos.SubscriptionEvent) error {
	if err := s.pubsub.Unsubscribe(unsubscribe); err != nil {
		return err
	}
	return nil
}

// LV new gateway functions
func (s *Service) SubscribeAux(subscription *protos.SubscriptionEvent) error {
	return s.pubsub.SubscribeAux(subscription)
}


func (s *Service) SubscribeStreamAux(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamServer) error {
	return s.pubsub.SubscribeStreamAux(subscription, stream)
}

func (s *Service) UnsubscribeAux(unsubscribe *protos.SubscriptionEvent) error {
	if err := s.pubsub.UnsubscribeAux(unsubscribe); err != nil {
		return err
	}
	return nil
}


// LV END



//LV new gateway functions
func (s *Service) RegisterMDSGateway(gateway *protos.GatewayConfig) error {
	if err := s.kvStore.RegisterMDSGateway(gateway); err != nil {
		return err
	}
	return nil
}


func (s *Service) ListMDSGateways() (*protos.AvailableGatewayConfigs, error) {
	return s.kvStore.ListMDSGateways()
}

// LV END