package mdsservice

import (
	"context"
	"io"

	"time"

	"github.com/IBM/gedsmds/internal/logger"
	"github.com/IBM/gedsmds/internal/mdsprocessor"
	"github.com/IBM/gedsmds/internal/prommetrics"
	"github.com/IBM/gedsmds/protos"
)

func NewService(metrics *prommetrics.Metrics) *Service {
	return &Service{
		processor: mdsprocessor.InitService(),
		metrics:   metrics,
	}
}

func (s *Service) GetConnectionInformation(ctx context.Context,
	_ *protos.EmptyParams) (*protos.ConnectionInformation, error) {
	if address, err := s.processor.GetClientConnectionInformation(ctx); err != nil {
		return &protos.ConnectionInformation{}, err
	} else {
		logger.InfoLogger.Println("sending connection info: ", address)
		return &protos.ConnectionInformation{RemoteAddress: address}, nil
	}
}

// LV new definitions - helper function for tracing
func (s *Service) GetClientConnAux(ctx context.Context) (*protos.ConnectionInformation) {
	if address, err := s.processor.GetClientConnectionInformation(ctx); err != nil {
		return nil
	} else {
		return &protos.ConnectionInformation{RemoteAddress: address}
	}
}
// LV END



func (s *Service) RegisterObjectStore(ctx context.Context,
	objectStore *protos.ObjectStoreConfig) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("register objectStore", objectStore)
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";GEN;REG_STORE;", objectStore.Bucket, ";", objectStore) 
	if err := s.processor.RegisterObjectStore(objectStore); err != nil {
		// It should return status code already exist
		return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) ListObjectStores(ctx context.Context,
	_ *protos.EmptyParams) (*protos.AvailableObjectStoreConfigs, error) {
	logger.InfoLogger.Println("list object stores")
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";GEN;LST_STORE") 
	return s.processor.ListObjectStores()
}

func (s *Service) CreateBucket(ctx context.Context, bucket *protos.Bucket) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("create bucket", bucket)
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_CRT;", bucket.Bucket) 
	if err := s.processor.CreateBucket(bucket); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_ALREADY_EXISTS}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) DeleteBucket(ctx context.Context, bucket *protos.Bucket) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("delete bucket", bucket)
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_DEL;", bucket.Bucket) 
	if err := s.processor.DeleteBucket(bucket); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) ListBuckets(ctx context.Context, _ *protos.EmptyParams) (*protos.BucketListResponse, error) {
	logger.InfoLogger.Println("list buckets")
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LST") 
	aux, err := s.processor.ListBuckets()
	for _, bucket := range aux.Results {
		logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LST;", bucket) 
	}
	return aux, err
	// return s.processor.ListBuckets()

}

func (s *Service) LookupBucket(ctx context.Context, bucket *protos.Bucket) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("look up bucket", bucket)
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LKP;", bucket.Bucket) 
	if err := s.processor.LookupBucket(bucket); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) Create(ctx context.Context, object *protos.Object) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("create object", object)
	s.metrics.IncrementCreateObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";OBJ;OBJ_CRT;", object.Id.Bucket, ";", object.Id.Key ,";", object.Info.Location, ";", object.Info.Size, ";", object.Info.SealedOffset)
	if err := s.processor.CreateObject(object); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_ALREADY_EXISTS}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) Update(ctx context.Context, object *protos.Object) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("update object", object)
	s.metrics.IncrementUpdateObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";OBJ;OBJ_UPD;", object.Id.Bucket, ";", object.Id.Key ,";", object.Info.Location, ";", object.Info.Size, ";", object.Info.SealedOffset)
	if err := s.processor.UpdateObject(object); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_INTERNAL}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) Delete(ctx context.Context, objectID *protos.ObjectID) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("delete object", objectID)
	s.metrics.IncrementDeleteObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";OBJ;OBJ_DEL;", objectID.Bucket, ";", objectID.Key)	
	if err := s.processor.DeleteObject(objectID); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) DeletePrefix(ctx context.Context, objectID *protos.ObjectID) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("delete object", objectID)
	s.metrics.IncrementDeleteObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PFX;PFX_DEL;", objectID.Bucket, ";", objectID.Key) 
	if err := s.processor.DeletePrefix(objectID); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) Lookup(ctx context.Context, objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	logger.InfoLogger.Println("lookup object", objectID)
	s.metrics.IncrementLookupObject()
	// LV TRACE
	logger.TraceLogger.Print(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";OBJ;OBJ_LKP;", objectID.Bucket, ";", objectID.Key ,";")
	object, err := s.processor.LookupObject(objectID)
	if err != nil {
		logger.TraceLogger.Println(object.Result.Info.Location, ";", object.Result.Info.Size, ";", object.Result.Info.SealedOffset)
		return &protos.ObjectResponse{
			Error: &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND},
		}, err
	}
	return object, nil
}

func (s *Service) List(ctx context.Context, objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	logger.InfoLogger.Println("list objects")
	s.metrics.IncrementListObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PFX;PFX_LKP;", objectListRequest.Prefix)
	return s.processor.List(objectListRequest)
}

func (s *Service) CreateOrUpdateObjectStream(stream protos.MetadataService_CreateOrUpdateObjectStreamServer) error {
	for {
		object, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.StatusResponse{Code: protos.StatusCode_OK})
		}
		if err != nil {
			return err
		}
		s.processor.CreateOrUpdateObjectStream(object)
	}
}

func (s *Service) Subscribe(ctx context.Context, subscription *protos.SubscriptionEvent) (*protos.StatusResponse, error) {
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PUB;PUB_SEV;", subscription.SubscriptionType, ";", subscription.SubscriberID,";", subscription.BucketID,";", subscription.Key) 
	if err := s.processor.Subscribe(subscription); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_INTERNAL}, err
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) SubscribeStream(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamServer) error {
	// LV TRACE
	logger.TraceLogger.Println("----;", time.Now().UnixMicro(),";PUB;PUB_SST;----;", subscription.SubscriberID) 
	return s.processor.SubscribeStream(subscription, stream)
}

func (s *Service) Unsubscribe(ctx context.Context, unsubscribe *protos.SubscriptionEvent) (*protos.StatusResponse, error) {
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PUB;PUB_UNS;", unsubscribe.SubscriptionType, ";", unsubscribe.SubscriberID,";", unsubscribe.BucketID,";", unsubscribe.Key) 
	if err := s.processor.Unsubscribe(unsubscribe); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

//LV New operations for MDS-gateway federation

func (s *Service) RegisterMDSGateway(_ context.Context,
	gateway *protos.GatewayConfig) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("register MDS gateway", gateway)
	if err := s.processor.RegisterMDSGateway(gateway); err != nil {
		// It should return status code already exist
		return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) ListMDSGateways(_ context.Context,
	_ *protos.EmptyParams) (*protos.AvailableGatewayConfigs, error) {
	logger.InfoLogger.Println("list MDS gateways")
	return s.processor.ListMDSGateways()
}

func (s *Service) LookupObjectAux(ctx context.Context, objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	logger.InfoLogger.Println("LookupObjectAux metadata called from external gateway", objectID)
	s.metrics.IncrementLookupObject()
	// LV TRACE
	logger.TraceLogger.Print(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";OBJ;OBJ_LKP_AUX;", objectID.Bucket, ";", objectID.Key ,";")
	object, err := s.processor.LookupObjectAux(objectID)
	logger.InfoLogger.Println("LookupObjectAux metadata returned to external gateway", object)
	if err != nil {
		return &protos.ObjectResponse{
			Error: &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND},
		}, err
	}
	return object, nil
}

func (s *Service) LookupBucketAux(ctx context.Context, bucket *protos.Bucket) (*protos.StatusResponse, error) {
	logger.InfoLogger.Println("LookupBUCKETAux", bucket)
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LKP_AUX;", bucket.Bucket) 
	if err := s.processor.LookupBucketAux(bucket); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) ListBucketsAux(ctx context.Context, _ *protos.EmptyParams) (*protos.BucketListResponse, error) {
	logger.InfoLogger.Println("list buckets")
	
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LST_AUX") 
	aux, err := s.processor.ListBucketsAux()
	for _, bucket := range aux.Results {
		logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";BKT;BKT_LST_AUX;", bucket) 
	}
	return aux, err
}


func (s *Service) ListAux(ctx context.Context, objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	logger.InfoLogger.Println("ListObjectsAux")
	s.metrics.IncrementListObject()
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PFX;PFX_LKP_AUX;", objectListRequest.Prefix)
	return s.processor.ListAux(objectListRequest)
}

func (s *Service) SubscribeAux(ctx context.Context, subscription *protos.SubscriptionEvent) (*protos.StatusResponse, error) {
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PUB;PUB_SEV_AUX;", subscription.SubscriptionType, ";", subscription.SubscriberID,";", subscription.BucketID,";", subscription.Key) 
	if err := s.processor.SubscribeAux(subscription); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_INTERNAL}, err
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}

func (s *Service) SubscribeStreamAux(subscription *protos.SubscriptionStreamEvent,
	stream protos.MetadataService_SubscribeStreamAuxServer) error {
	// LV TRACE
	logger.TraceLogger.Println("----;", time.Now().UnixMicro(),";PUB;PUB_SST_AUX;----;", subscription.SubscriberID) 	
	return s.processor.SubscribeStreamAux(subscription, stream)
}

func (s *Service) UnsubscribeAux(ctx context.Context, unsubscribe *protos.SubscriptionEvent) (*protos.StatusResponse, error) {
	// LV TRACE
	logger.TraceLogger.Println(s.GetClientConnAux(ctx).RemoteAddress, ";", time.Now().UnixMicro(),";PUB;PUB_UNS_AUX;", unsubscribe.SubscriptionType, ";", unsubscribe.SubscriberID,";", unsubscribe.BucketID,";", unsubscribe.Key) 
	if err := s.processor.UnsubscribeAux(unsubscribe); err != nil {
		return &protos.StatusResponse{Code: protos.StatusCode_NOT_FOUND}, nil
	}
	return &protos.StatusResponse{Code: protos.StatusCode_OK}, nil
}


// LV END