package keyvaluestore

import (
	"errors"
	"github.com/IBM/gedsmds/internal/config"
	"github.com/IBM/gedsmds/internal/keyvaluestore/db"
	"github.com/IBM/gedsmds/internal/logger"
	"github.com/IBM/gedsmds/protos"
	"github.com/IBM/gedsmds/internal/connection/gtconnpool"
	"golang.org/x/exp/maps"
	"strings"
	"sync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
	"context"
)

//LV new definitions
type Executor struct {
	mdsGatewayConnections map[string]*gtconnpool.Pool
}

func NewExecutor(gateway string) *Executor {
	tempEx := &Executor{
		mdsGatewayConnections: gtconnpool.GetMDSGatewayConnectionsStream(gateway),
	}
	gtconnpool.SleepAndContinue()
	return tempEx
}
// LV END

//LV new defitions 
func (kv *Service) GatewayConfigs() (map[string]*protos.GatewayConfig){
	return kv.gatewayConfigs
}

func (kv *Service) GatewayConfigsLock() (*sync.RWMutex) {
	return kv.gatewayConfigsLock
}
//LV END


func InitKeyValueStoreService() *Service {
	kvStore := &Service{
		//dbConnection: db.NewOperations(),

		objectStoreConfigsLock: &sync.RWMutex{},
		objectStoreConfigs:     map[string]*protos.ObjectStoreConfig{},

		bucketsLock: &sync.RWMutex{},
		buckets:     map[string]*Bucket{},

		//LV new definitions
		gatewayConfigsLock: &sync.RWMutex{},
		gatewayConfigs:     map[string]*protos.GatewayConfig{},
		//LV end
	}
	go kvStore.populateCache()
	return kvStore
}

func (kv *Service) NewBucketIfNotExist(objectID *protos.ObjectID) {
	if _, ok := kv.buckets[objectID.Bucket]; !ok {
		kv.buckets[objectID.Bucket] = &Bucket{
			bucket:      &protos.Bucket{Bucket: objectID.Bucket},
			objectsLock: &sync.RWMutex{},
			objects:     map[string]*Object{},
		}
	}
}

func (kv *Service) populateCache() {
	if config.Config.PersistentStorageEnabled && config.Config.RepopulateCacheEnabled {
		logger.InfoLogger.Println("populating the cache")
		go kv.populateObjectStoreConfig()
		go kv.populateBuckets()
	}
}

func (kv *Service) populateObjectStoreConfig() {
/*	if allObjectStoreConfig, err := kv.dbConnection.GetAllObjectStoreConfig(); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		kv.objectStoreConfigsLock.Lock()
		for _, objectStoreConfig := range allObjectStoreConfig {
			kv.objectStoreConfigs[objectStoreConfig.Bucket] = objectStoreConfig
		}
		kv.objectStoreConfigsLock.Unlock()
	}
	*/
}

func (kv *Service) populateBuckets() {
/*	if allBuckets, err := kv.dbConnection.GetAllBuckets(); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		kv.bucketsLock.Lock()
		for _, bucket := range allBuckets {
			kv.NewBucketIfNotExist(&protos.ObjectID{Bucket: bucket.Bucket})
		}
		kv.bucketsLock.Unlock()
	}
	if allObjects, err := kv.dbConnection.GetAllObjects(); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		kv.bucketsLock.Lock()
		for _, object := range allObjects {
			kv.NewBucketIfNotExist(object.Id)
			kv.buckets[object.Id.Bucket].objects[object.Id.Key] = &Object{object: object, path: kv.getNestedPath(object.Id)}
		}
		kv.bucketsLock.Unlock()
	}
*/

}

func (kv *Service) RegisterObjectStore(objectStore *protos.ObjectStoreConfig) error {
	kv.objectStoreConfigsLock.Lock()
	defer kv.objectStoreConfigsLock.Unlock()
	if _, ok := kv.objectStoreConfigs[objectStore.Bucket]; ok {
		return errors.New("config already exists")
	}
	kv.objectStoreConfigs[objectStore.Bucket] = objectStore
	if config.Config.PersistentStorageEnabled {
		//kv.dbConnection.ObjectStoreConfigChan <- &db.OperationParams{
		//	ObjectStoreConfig: objectStore,
		//	Type:              db.PUT,
		//}
	}
	return nil
}

func (kv *Service) ListObjectStores() (*protos.AvailableObjectStoreConfigs, error) {
	kv.objectStoreConfigsLock.RLock()
	defer kv.objectStoreConfigsLock.RUnlock()
	mappings := &protos.AvailableObjectStoreConfigs{Mappings: []*protos.ObjectStoreConfig{}}
	for _, objectStoreConfig := range kv.objectStoreConfigs {
		mappings.Mappings = append(mappings.Mappings, objectStoreConfig)
		logger.TraceLogger.Println("------------;", time.Now().UnixMicro(),";GEN;LST_STORE;", objectStoreConfig.Bucket, ";", objectStoreConfig) 
	}
	return mappings, nil
}


// LV new definitions

func (kv *Service) RegisterMDSGateway(gateway *protos.GatewayConfig) error {
	kv.gatewayConfigsLock.Lock()
	defer kv.gatewayConfigsLock.Unlock()
	if _, ok := kv.gatewayConfigs[gateway.RemoteAddress]; ok {
		return errors.New("gateway already exists")
	}
	kv.gatewayConfigs[gateway.RemoteAddress] = gateway
	NewExecutor(gateway.RemoteAddress) 
		
	return nil
}

func (kv *Service) ListMDSGateways() (*protos.AvailableGatewayConfigs, error) {
	kv.gatewayConfigsLock.RLock()
	defer kv.gatewayConfigsLock.RUnlock()
	mappings := &protos.AvailableGatewayConfigs{Mappings: []*protos.GatewayConfig{}}
	for _, gatewayConfig := range kv.gatewayConfigs {
		mappings.Mappings = append(mappings.Mappings, gatewayConfig)
	}
	return mappings, nil
}

// LV END


func (kv *Service) CreateBucket(bucket *protos.Bucket) error {
	kv.bucketsLock.Lock()
	defer kv.bucketsLock.Unlock()
	if _, ok := kv.buckets[bucket.Bucket]; ok {
		return errors.New("bucket already exists")
	}
	kv.NewBucketIfNotExist(&protos.ObjectID{Bucket: bucket.Bucket})
	if config.Config.PersistentStorageEnabled {
		kv.dbConnection.BucketChan <- &db.OperationParams{
			Bucket: bucket,
			Type:   db.PUT,
		}
	}
	return nil
}

func (kv *Service) DeleteBucket(bucket *protos.Bucket) error {
	kv.bucketsLock.Lock()
	defer kv.bucketsLock.Unlock()
	if _, ok := kv.buckets[bucket.Bucket]; !ok {
		return errors.New("bucket already deleted")
	}
	delete(kv.buckets, bucket.Bucket)
	if config.Config.PersistentStorageEnabled {
		// delete the bucket and all objects in it.
		kv.dbConnection.BucketChan <- &db.OperationParams{
			Bucket: bucket,
			Type:   db.DELETE,
		}
	}
	return nil
}

func (kv *Service) ListBuckets() (*protos.BucketListResponse, error) {
	/*kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	buckets := &protos.BucketListResponse{Results: []string{}}
	buckets.Results = append(buckets.Results, maps.Keys(kv.buckets)...)
	return buckets, nil
	*/
	//LV : lookup prefix locally and return it if found
	logger.InfoLogger.Println("first perform local metadata bucket list")
	
	result, err  := kv.ListBucketsAux() 

	if (err == nil) {
		logger.InfoLogger.Println("local metadata bucket list found: ", result, err)
		//return result, nil 
	}

	//LV: iterate gateways until object is found (first one found is returned, assume namespace partitioning by owner/user)
	logger.ErrorLogger.Println("list buckets - iterating connections to known gateways")
	for _, gatewayConfig := range kv.gatewayConfigs {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		conn, err := factoryNode(gatewayConfig.RemoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway ", gatewayConfig.RemoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway", gatewayConfig.RemoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		emptyParams := &protos.EmptyParams{}
		result2, err := client.ListBucketsAux(context.Background(), emptyParams)
		if err != nil {
			logger.ErrorLogger.Println(err)
		} else {
			logger.InfoLogger.Println("list buckets - metadata returned from gateway", gatewayConfig, result2)
			for _, result2_aux := range result2.Results {
				result.Results = append(result.Results, result2_aux)
			}
		}
	}
	return result, err
}

// LV - new definitions 

func (kv *Service) LookupBucketAux(bucket *protos.Bucket) error {
	kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	if _, ok := kv.buckets[bucket.Bucket]; !ok {
		logger.ErrorLogger.Println("Bucket", bucket.Bucket, "NOT found locally")
		return errors.New("bucket does not exist")
	}
	logger.InfoLogger.Println("Bucket", bucket.Bucket, "found locally")
	return nil
}

func (kv *Service) LookupBucket(bucket *protos.Bucket) error {
	//LV : lookup bucket locally and return it if found
	logger.InfoLogger.Println("first perform local metadata BUCKET lookup", bucket)
	result := kv.LookupBucketAux(bucket) 
	if result == nil {
		logger.InfoLogger.Println("returning local metadata BUCKET lookup", result)
		logger.InfoLogger.Println("BUCKET metadata returned finally",  result)
		return nil 
	}

	//LV: iterate gateways until BUCKET is found (first one found is returned, assume namespace partitioning by owner/user)
	kv.gatewayConfigsLock.RLock()
	defer kv.gatewayConfigsLock.RUnlock()
	
	logger.ErrorLogger.Println("BUCKET not found locally - iterating connections to known gateways")
	for _, gatewayConfig := range kv.gatewayConfigs {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		conn, err := factoryNode(gatewayConfig.RemoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway ", gatewayConfig.RemoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway", gatewayConfig.RemoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		bucket := &protos.Bucket{
		Bucket:    bucket.Bucket,
		}
		
		result, err := client.LookupBucketAux(context.Background(), bucket)
		if err == nil {
			logger.InfoLogger.Println("BUCKET metadata returned from gateway", gatewayConfig, result.Code)
			logger.InfoLogger.Println("BUCKET metadata returned finally",  result.Code)
			return nil //LV - should be  return result?
		}
		/*if err != nil {
			logger.ErrorLogger.Println(err)
		} else {
			logger.InfoLogger.Println("BUCKET metadata returned from gateway", gatewayConfig, result)
			logger.InfoLogger.Println("BUCKET metadata returned finally",  result)
			res2 := result 
			return res2 //LV - should be  return result
		}*/
	}
	logger.ErrorLogger.Println("Bucket", bucket.Bucket, "NOT found locally NEITHER in any gateway")
	return errors.New("bucket does not exist")
}	
// LV -end

func (kv *Service) LookupBucketByName(bucketName string) error {
	kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	if _, ok := kv.buckets[bucketName]; !ok {
		return errors.New("bucket does not exist")
	}
	return nil
}

func (kv *Service) CreateObject(object *protos.Object) error {
	newObject := &Object{object: object, path: kv.getNestedPath(object.Id)}
	kv.bucketsLock.Lock()
	kv.NewBucketIfNotExist(object.Id)
	kv.buckets[object.Id.Bucket].objects[object.Id.Key] = newObject
	kv.bucketsLock.Unlock()
	logger.InfoLogger.Println(object)
	if config.Config.PersistentStorageEnabled {
		kv.dbConnection.ObjectChan <- &db.OperationParams{
			Object: object,
			Type:   db.PUT,
		}
	}
	return nil
}

func (kv *Service) UpdateObject(object *protos.Object) error {
	newObject := &Object{object: object, path: kv.getNestedPath(object.Id)}
	kv.bucketsLock.Lock()
	kv.NewBucketIfNotExist(object.Id)
	kv.buckets[object.Id.Bucket].objects[object.Id.Key] = newObject
	kv.bucketsLock.Unlock()
	if config.Config.PersistentStorageEnabled {
		kv.dbConnection.ObjectChan <- &db.OperationParams{
			Object: object,
			Type:   db.PUT,
		}
	}
	return nil
}

func (kv *Service) CreateOrUpdateObject(object *protos.Object) error {
	newObject := &Object{object: object, path: kv.getNestedPath(object.Id)}
	kv.bucketsLock.Lock()
	kv.NewBucketIfNotExist(object.Id)
	kv.buckets[object.Id.Bucket].objects[object.Id.Key] = newObject
	kv.bucketsLock.Unlock()
	if config.Config.PersistentStorageEnabled {
		kv.dbConnection.ObjectChan <- &db.OperationParams{
			Object: object,
			Type:   db.PUT,
		}
	}
	return nil
}

func (kv *Service) DeleteObject(objectID *protos.ObjectID) error {
	kv.bucketsLock.Lock()
	if _, ok := kv.buckets[objectID.Bucket]; ok {
		delete(kv.buckets[objectID.Bucket].objects, objectID.Key)
	}
	kv.bucketsLock.Unlock()
	if config.Config.PersistentStorageEnabled {
		kv.dbConnection.ObjectChan <- &db.OperationParams{
			Object: &protos.Object{
				Id: objectID,
			},
			Type: db.DELETE,
		}
	}
	return nil
}

func (kv *Service) DeleteObjectPrefix(objectID *protos.ObjectID) ([]*protos.Object, error) {
	var deletedObjects []*protos.Object
	// this will be slow, needs to be optimized
	kv.bucketsLock.Lock()
	if _, ok := kv.buckets[objectID.Bucket]; ok {
		for key, object := range kv.buckets[objectID.Bucket].objects {
			if strings.HasPrefix(key, objectID.Key) {
				deletedObjects = append(deletedObjects, object.object)
				delete(kv.buckets[objectID.Bucket].objects, objectID.Key)
			}
		}
	}
	kv.bucketsLock.Unlock()
	if config.Config.PersistentStorageEnabled {
		for _, object := range deletedObjects {
			kv.dbConnection.ObjectChan <- &db.OperationParams{
				Object: object,
				Type:   db.DELETE,
			}
		}
	}
	return deletedObjects, nil
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

func (kv *Service) LookupObjectAux(objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	if kv.buckets[objectID.Bucket] == nil {
		return nil, errors.New("bucket's object does not exist")
	}
	if _, ok := kv.buckets[objectID.Bucket].objects[objectID.Key]; !ok {
		return nil, errors.New("object does not exist")
	}
	return &protos.ObjectResponse{
		Result: kv.buckets[objectID.Bucket].objects[objectID.Key].object,
	}, nil
}


func (kv *Service) LookupObject(objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	//LV : lookup object locally and return it if found
	logger.InfoLogger.Println("first perform local metadata object lookup", objectID)
	result, err  := kv.LookupObjectAux(objectID) 
	if (err == nil) {
		logger.InfoLogger.Println("returning local metadata object lookup", result, err)
		return result, nil 
	}
	
	kv.gatewayConfigsLock.RLock()
	defer kv.gatewayConfigsLock.RUnlock()
		
	//LV: iterate gateways until object is found (first one found is returned, assume namespace partitioning by owner/user)
	logger.ErrorLogger.Println("object not found locally - iterating connections to known gateways")
	for _, gatewayConfig := range kv.gatewayConfigs {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		conn, err := factoryNode(gatewayConfig.RemoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway ", gatewayConfig.RemoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway", gatewayConfig.RemoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		objectId := &protos.ObjectID{
		Key:    objectID.Key,
		Bucket: objectID.Bucket,
		}
	
		result, err := client.LookupObjectAux(context.Background(), objectId)
		if err != nil {
			logger.ErrorLogger.Println(err)
		} else {
			logger.InfoLogger.Println("metadata returned from gateway", gatewayConfig, result.Result)
			if (result.Result != nil){
				return result, err
			}
		}
	}
	return result, err
}	

func (kv *Service) ListObjectsAux(objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	objects := &protos.ObjectListResponse{Results: []*protos.Object{}, CommonPrefixes: []string{}}
	if objectListRequest.Prefix == nil || len(objectListRequest.Prefix.Bucket) == 0 {
		logger.InfoLogger.Println("bucket not set")
		return objects, nil
	}
	var delimiter string
	if objectListRequest.Delimiter != nil && *objectListRequest.Delimiter != 0 {
		delimiter = string(*objectListRequest.Delimiter)
	}
	tempCommonPrefixes := map[string]bool{}
	// needs to be optimized
	if len(delimiter) == 0 {
		kv.bucketsLock.RLock()
		if _, ok := kv.buckets[objectListRequest.Prefix.Bucket]; ok {
			for key, object := range kv.buckets[objectListRequest.Prefix.Bucket].objects {
				if strings.HasPrefix(key, objectListRequest.Prefix.Key) {
					objects.Results = append(objects.Results, object.object)
				}
			}
		}
		kv.bucketsLock.RUnlock()
	} else {
		if len(objectListRequest.Prefix.Key) == 0 {
			kv.bucketsLock.RLock()
			if _, ok := kv.buckets[objectListRequest.Prefix.Bucket]; ok {
				for _, object := range kv.buckets[objectListRequest.Prefix.Bucket].objects {
					if len(object.path) == 1 {
						objects.Results = append(objects.Results, object.object)
					} else if len(object.path) > 1 {
						tempCommonPrefixes[object.path[0]] = true
					}
				}
			}
			kv.bucketsLock.RUnlock()
		} else {
			prefixLength := len(strings.Split(objectListRequest.Prefix.Key, delimiter)) + 1
			kv.bucketsLock.RLock()
			if _, ok := kv.buckets[objectListRequest.Prefix.Bucket]; ok {
				for key, object := range kv.buckets[objectListRequest.Prefix.Bucket].objects {
					if strings.HasPrefix(key, objectListRequest.Prefix.Key) {
						objects.Results = append(objects.Results, object.object)
						if len(object.path) == prefixLength {
							tempCommonPrefixes[strings.Join(object.path[:prefixLength-1], delimiter)] = true
						}
					}
				}
			}
			kv.bucketsLock.RUnlock()
		}
	}
	if len(tempCommonPrefixes) > 0 {
		for commonPrefix := range tempCommonPrefixes {
			objects.CommonPrefixes = append(objects.CommonPrefixes, commonPrefix+commonDelimiter)
		}
	}

	return objects, nil
}

func (kv *Service) ListObjects(objectListRequest *protos.ObjectListRequest) (*protos.ObjectListResponse, error) {
	//LV : list objects locally and save it if found
	logger.InfoLogger.Println("first perform local object list", objectListRequest)
	
	result, err  := kv.ListObjectsAux(objectListRequest) 
	
	if (err == nil) {
		logger.InfoLogger.Println("local metadata local object list found: ", result, err)
		//return result, nil 
	}
	
	kv.gatewayConfigsLock.RLock()
	defer kv.gatewayConfigsLock.RUnlock()
		
	//LV: iterate gateways until object is found (first one found is returned, assume namespace partitioning by owner/user)
	logger.ErrorLogger.Println("listobjects - iterating connections to known gateways")
	for _, gatewayConfig := range kv.gatewayConfigs {
		//conn, err := gtconnpool.GetMDSGatewayConnectionsStream(gatewayConfig.RemoteAddress)[gatewayConfig.RemoteAddress].clients
		conn, err := factoryNode(gatewayConfig.RemoteAddress)
		if conn == nil || err != nil {
			logger.ErrorLogger.Println("Error connection to MDSGategway ", gatewayConfig.RemoteAddress)
		}
		logger.InfoLogger.Println("Setup connection to gateway", gatewayConfig.RemoteAddress)
		client := protos.NewMetadataServiceClient(conn)
		
		//var objectListRequest protos.ObjectListRequest
		
		/*if objectListRequest.Delimiter != nil && *objectListRequest.Delimiter != 0 {
			var delimiter = string(*objectListRequest.Delimiter)
			objectListRequest := &protos.ObjectListRequest{
				Prefix: objectListRequest.Prefix,
				Delimeter: delimeter,
			}
		} else {
			*/
			objectListRequest := &protos.ObjectListRequest{
				Prefix: objectListRequest.Prefix,
			}
	
		/*}*/
	
		result2, err := client.ListAux(context.Background(), objectListRequest)
		if err != nil {
			logger.ErrorLogger.Println(err)
		} else {
			logger.InfoLogger.Println("listobjects - metadata returned from gateway", gatewayConfig, result2)
			for _, result2_aux := range result2.Results {
				result.Results = append(result.Results, result2_aux)
			}
		}
	}
	return result, err
}	


/*func (kv *Service) LookupObject(objectID *protos.ObjectID) (*protos.ObjectResponse, error) {
	kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	if kv.buckets[objectID.Bucket] == nil {
		return nil, errors.New("bucket's object does not exist")
	}
	if _, ok := kv.buckets[objectID.Bucket].objects[objectID.Key]; !ok {
		return nil, errors.New("object does not exist")
	}
	return &protos.ObjectResponse{
		Result: kv.buckets[objectID.Bucket].objects[objectID.Key].object,
	}, nil
}*/



func (kv *Service) ListBucketsAux() (*protos.BucketListResponse, error) {
	kv.bucketsLock.RLock()
	defer kv.bucketsLock.RUnlock()
	buckets := &protos.BucketListResponse{Results: []string{}}
	buckets.Results = append(buckets.Results, maps.Keys(kv.buckets)...)
	return buckets, nil
}


// LV END

