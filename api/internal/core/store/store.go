/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/shiningrush/droplet/data"

	"github.com/apisix/manager-api/internal/core/entity"
	"github.com/apisix/manager-api/internal/core/storage"
	"github.com/apisix/manager-api/internal/log"
	"github.com/apisix/manager-api/internal/utils"
	"github.com/apisix/manager-api/internal/utils/runtime"
)

var (
	storeNeedReInit = make([]*GenericStore, 0)
)

type Pagination struct {
	PageSize   int `json:"page_size" form:"page_size" auto_read:"page_size"`
	PageNumber int `json:"page" form:"page" auto_read:"page"`
}

type Interface interface {
	Type() HubKey
	Get(ctx context.Context, key string) (any, error)
	List(ctx context.Context, input ListInput) (*ListOutput, error)
	Create(ctx context.Context, obj any) (any, error)
	Update(ctx context.Context, obj any, createIfNotExist bool) (any, error)
	BatchDelete(ctx context.Context, keys []string) error
}

type GenericStore struct {
	Stg      storage.Interface
	initLock sync.Mutex

	cache sync.Map
	opt   GenericStoreOption

	cancel  context.CancelFunc
	closing bool
}

type GenericStoreOption struct {
	BasePath   string
	ObjType    reflect.Type
	KeyFunc    func(obj any) string
	StockCheck func(obj any, stockObj any) error
	Validator  Validator
	HubKey     HubKey
}

func NewGenericStore(opt GenericStoreOption) (*GenericStore, error) {
	if opt.BasePath == "" {
		log.Error("base path empty")
		return nil, fmt.Errorf("base path can not be empty")
	}
	if opt.ObjType == nil {
		log.Errorf("object type is nil")
		return nil, fmt.Errorf("object type can not be nil")
	}
	if opt.KeyFunc == nil {
		log.Error("key func is nil")
		return nil, fmt.Errorf("key func can not be nil")
	}

	if opt.ObjType.Kind() == reflect.Ptr {
		opt.ObjType = opt.ObjType.Elem()
	}
	if opt.ObjType.Kind() != reflect.Struct {
		log.Error("obj type is invalid")
		return nil, fmt.Errorf("obj type is invalid")
	}
	s := &GenericStore{
		opt: opt,
	}
	s.Stg = storage.GenEtcdStorage()

	return s, nil
}

func ReInit() error {
	for _, store := range storeNeedReInit {
		if err := store.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (s *GenericStore) Init() error {
	s.initLock.Lock()
	defer s.initLock.Unlock()
	return s.listAndWatch()
}

func (s *GenericStore) Type() HubKey {
	return s.opt.HubKey
}

func (s *GenericStore) Get(_ context.Context, key string) (any, error) {
	ret, ok := s.cache.Load(key)
	if !ok {
		log.Warnf("data not found by key: %s", key)
		return nil, data.ErrNotFound
	}
	return ret, nil
}

type ListInput struct {
	Predicate func(obj any) bool
	Format    func(obj any) any
	PageSize  int
	// start from 1
	PageNumber int
	Less       func(i, j any) bool
}

type ListOutput struct {
	Rows      []any `json:"rows"`
	TotalSize int   `json:"total_size"`
}

// NewListOutput returns JSON marshalling safe struct pointer for empty slice
func NewListOutput() *ListOutput {
	return &ListOutput{Rows: make([]any, 0)}
}

var defLessFunc = func(i, j any) bool {
	iBase := i.(entity.GetBaseInfo).GetBaseInfo()
	jBase := j.(entity.GetBaseInfo).GetBaseInfo()
	if iBase.CreateTime != jBase.CreateTime {
		return iBase.CreateTime < jBase.CreateTime
	}
	if iBase.UpdateTime != jBase.UpdateTime {
		return iBase.UpdateTime < jBase.UpdateTime
	}
	iID := utils.InterfaceToString(iBase.ID)
	jID := utils.InterfaceToString(jBase.ID)
	return iID < jID
}

func (s *GenericStore) List(_ context.Context, input ListInput) (*ListOutput, error) {
	var ret []any
	s.cache.Range(func(key, value any) bool {
		if input.Predicate != nil && !input.Predicate(value) {
			return true
		}
		if input.Format != nil {
			value = input.Format(value)
		}
		ret = append(ret, value)
		return true
	})

	//should return an empty array not a null for client
	if ret == nil {
		ret = []any{}
	}

	output := &ListOutput{
		Rows:      ret,
		TotalSize: len(ret),
	}
	if input.Less == nil {
		input.Less = defLessFunc
	}

	sort.Slice(output.Rows, func(i, j int) bool {
		return input.Less(output.Rows[i], output.Rows[j])
	})

	if input.PageSize > 0 && input.PageNumber > 0 {
		skipCount := (input.PageNumber - 1) * input.PageSize
		if skipCount > output.TotalSize {
			output.Rows = []any{}
			return output, nil
		}

		endIdx := skipCount + input.PageSize
		if endIdx >= output.TotalSize {
			output.Rows = ret[skipCount:]
			return output, nil
		}
		output.Rows = ret[skipCount:endIdx]
	}

	return output, nil
}

func (s *GenericStore) Range(_ context.Context, f func(key string, obj any) bool) {
	s.cache.Range(func(key, value any) bool {
		return f(key.(string), value)
	})
}

func (s *GenericStore) ingestValidate(obj any) (err error) {
	if s.opt.Validator != nil {
		if err := s.opt.Validator.Validate(obj); err != nil {
			log.Errorf("data validate failed: %s, %v", err, obj)
			return err
		}
	}

	if s.opt.StockCheck != nil {
		s.cache.Range(func(key, value any) bool {
			if err = s.opt.StockCheck(obj, value); err != nil {
				return false
			}
			return true
		})
	}
	return err
}

func (s *GenericStore) CreateCheck(obj any) ([]byte, error) {

	if setter, ok := obj.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.Creating()
	}

	if err := s.ingestValidate(obj); err != nil {
		return nil, err
	}

	key := s.opt.KeyFunc(obj)
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	_, ok := s.cache.Load(key)
	if ok {
		log.Warnf("key: %s is conflicted", key)
		return nil, fmt.Errorf("key: %s is conflicted", key)
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("json marshal failed: %s", err)
		return nil, fmt.Errorf("json marshal failed: %s", err)
	}

	return bytes, nil
}

func (s *GenericStore) Create(ctx context.Context, obj any) (any, error) {
	if setter, ok := obj.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.Creating()
	}

	bytes, err := s.CreateCheck(obj)
	if err != nil {
		return nil, err
	}

	if err := s.Stg.Create(ctx, s.GetObjStorageKey(obj), string(bytes)); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *GenericStore) Update(ctx context.Context, obj any, createIfNotExist bool) (any, error) {
	if err := s.ingestValidate(obj); err != nil {
		return nil, err
	}

	key := s.opt.KeyFunc(obj)
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	storedObj, ok := s.cache.Load(key)
	if !ok {
		if createIfNotExist {
			return s.Create(ctx, obj)
		}
		log.Warnf("key: %s is not found", key)
		return nil, fmt.Errorf("key: %s is not found", key)
	}

	if setter, ok := obj.(entity.GetBaseInfo); ok {
		storedGetter := storedObj.(entity.GetBaseInfo)
		storedInfo := storedGetter.GetBaseInfo()
		info := setter.GetBaseInfo()
		info.Updating(storedInfo)
	}

	bs, err := json.Marshal(obj)
	if err != nil {
		log.Errorf("json marshal failed: %s", err)
		return nil, fmt.Errorf("json marshal failed: %s", err)
	}
	if err := s.Stg.Update(ctx, s.GetObjStorageKey(obj), string(bs)); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *GenericStore) BatchDelete(ctx context.Context, keys []string) error {
	var storageKeys []string
	for i := range keys {
		storageKeys = append(storageKeys, s.GetStorageKey(keys[i]))
	}

	return s.Stg.BatchDelete(ctx, storageKeys)
}

func (s *GenericStore) listAndWatch() error {
	lc, lcancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer lcancel()
	ret, err := s.Stg.List(lc, s.opt.BasePath)
	if err != nil {
		return err
	}
	for i := range ret {
		key := ret[i].Key[len(s.opt.BasePath)+1:]
		objPtr, err := s.StringToObjPtr(ret[i].Value, key)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error occurred while initializing logical store: %s, err: %v", s.opt.BasePath, err)
			return err
		}

		s.cache.Store(s.opt.KeyFunc(objPtr), objPtr)
	}

	// start watch
	s.cancel = s.watch()

	return nil
}

func (s *GenericStore) watch() context.CancelFunc {
	c, cancel := context.WithCancel(context.TODO())
	ch := s.Stg.Watch(c, s.opt.BasePath)
	go func() {
		defer func() {
			if !s.closing {
				log.Errorf("etcd watch exception closed, restarting: resource: %s", s.Type())
				storeNeedReInit = append(storeNeedReInit, s)
			}
		}()
		defer runtime.HandlePanic()
		for event := range ch {
			if event.Canceled {
				log.Warnf("etcd watch failed: %s", event.Error)
				return
			}

			for i := range event.Events {
				switch event.Events[i].Type {
				case storage.EventTypePut:
					key := event.Events[i].Key[len(s.opt.BasePath)+1:]
					objPtr, err := s.StringToObjPtr(event.Events[i].Value, key)
					if err != nil {
						log.Warnf("value convert to obj failed: %s", err)
						continue
					}
					s.cache.Store(key, objPtr)
				case storage.EventTypeDelete:
					s.cache.Delete(event.Events[i].Key[len(s.opt.BasePath)+1:])
				}
			}
		}
	}()
	return cancel
}

func (s *GenericStore) Close() error {
	s.closing = true
	s.cancel()
	return nil
}

func (s *GenericStore) StringToObjPtr(str, key string) (any, error) {
	objPtr := reflect.New(s.opt.ObjType)
	ret := objPtr.Interface()
	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		log.Errorf("json unmarshal failed: %s", err)
		return nil, fmt.Errorf("json unmarshal failed\n\tRelated Key:\t\t%s\n\tError Description:\t%s", key, err)
	}

	if setter, ok := ret.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.KeyCompat(key)
	}

	return ret, nil
}

func (s *GenericStore) GetObjStorageKey(obj any) string {
	return s.GetStorageKey(s.opt.KeyFunc(obj))
}

func (s *GenericStore) GetStorageKey(key string) string {
	return fmt.Sprintf("%s/%s", s.opt.BasePath, key)
}
