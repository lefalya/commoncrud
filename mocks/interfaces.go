// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces/main.go

// Package mock_interfaces is a generated GoMock package.
package mock_interfaces

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	interfaces "github.com/lefalya/commoncrud/interfaces"
	commonlogger "github.com/lefalya/commonlogger"
	bson "go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"
)

// MockItem is a mock of Item interface.
type MockItem struct {
	ctrl     *gomock.Controller
	recorder *MockItemMockRecorder
}

// MockItemMockRecorder is the mock recorder for MockItem.
type MockItemMockRecorder struct {
	mock *MockItem
}

// NewMockItem creates a new mock instance.
func NewMockItem(ctrl *gomock.Controller) *MockItem {
	mock := &MockItem{ctrl: ctrl}
	mock.recorder = &MockItemMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockItem) EXPECT() *MockItemMockRecorder {
	return m.recorder
}

// GetCreatedAt mocks base method.
func (m *MockItem) GetCreatedAt() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreatedAt")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetCreatedAt indicates an expected call of GetCreatedAt.
func (mr *MockItemMockRecorder) GetCreatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreatedAt", reflect.TypeOf((*MockItem)(nil).GetCreatedAt))
}

// GetCreatedAtString mocks base method.
func (m *MockItem) GetCreatedAtString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreatedAtString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCreatedAtString indicates an expected call of GetCreatedAtString.
func (mr *MockItemMockRecorder) GetCreatedAtString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreatedAtString", reflect.TypeOf((*MockItem)(nil).GetCreatedAtString))
}

// GetRandId mocks base method.
func (m *MockItem) GetRandId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRandId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRandId indicates an expected call of GetRandId.
func (mr *MockItemMockRecorder) GetRandId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRandId", reflect.TypeOf((*MockItem)(nil).GetRandId))
}

// GetUUID mocks base method.
func (m *MockItem) GetUUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetUUID indicates an expected call of GetUUID.
func (mr *MockItemMockRecorder) GetUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUUID", reflect.TypeOf((*MockItem)(nil).GetUUID))
}

// GetUpdatedAt mocks base method.
func (m *MockItem) GetUpdatedAt() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatedAt")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetUpdatedAt indicates an expected call of GetUpdatedAt.
func (mr *MockItemMockRecorder) GetUpdatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatedAt", reflect.TypeOf((*MockItem)(nil).GetUpdatedAt))
}

// GetUpdatedAtString mocks base method.
func (m *MockItem) GetUpdatedAtString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatedAtString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetUpdatedAtString indicates an expected call of GetUpdatedAtString.
func (mr *MockItemMockRecorder) GetUpdatedAtString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatedAtString", reflect.TypeOf((*MockItem)(nil).GetUpdatedAtString))
}

// SetCreatedAt mocks base method.
func (m *MockItem) SetCreatedAt(time time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCreatedAt", time)
}

// SetCreatedAt indicates an expected call of SetCreatedAt.
func (mr *MockItemMockRecorder) SetCreatedAt(time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCreatedAt", reflect.TypeOf((*MockItem)(nil).SetCreatedAt), time)
}

// SetCreatedAtString mocks base method.
func (m *MockItem) SetCreatedAtString(timeString string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCreatedAtString", timeString)
}

// SetCreatedAtString indicates an expected call of SetCreatedAtString.
func (mr *MockItemMockRecorder) SetCreatedAtString(timeString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCreatedAtString", reflect.TypeOf((*MockItem)(nil).SetCreatedAtString), timeString)
}

// SetRandId mocks base method.
func (m *MockItem) SetRandId() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetRandId")
}

// SetRandId indicates an expected call of SetRandId.
func (mr *MockItemMockRecorder) SetRandId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRandId", reflect.TypeOf((*MockItem)(nil).SetRandId))
}

// SetUUID mocks base method.
func (m *MockItem) SetUUID() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUUID")
}

// SetUUID indicates an expected call of SetUUID.
func (mr *MockItemMockRecorder) SetUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUUID", reflect.TypeOf((*MockItem)(nil).SetUUID))
}

// SetUpdatedAt mocks base method.
func (m *MockItem) SetUpdatedAt(time time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdatedAt", time)
}

// SetUpdatedAt indicates an expected call of SetUpdatedAt.
func (mr *MockItemMockRecorder) SetUpdatedAt(time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdatedAt", reflect.TypeOf((*MockItem)(nil).SetUpdatedAt), time)
}

// SetUpdatedAtString mocks base method.
func (m *MockItem) SetUpdatedAtString(timeString string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdatedAtString", timeString)
}

// SetUpdatedAtString indicates an expected call of SetUpdatedAtString.
func (mr *MockItemMockRecorder) SetUpdatedAtString(timeString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdatedAtString", reflect.TypeOf((*MockItem)(nil).SetUpdatedAtString), timeString)
}

// MockMongoItem is a mock of MongoItem interface.
type MockMongoItem struct {
	ctrl     *gomock.Controller
	recorder *MockMongoItemMockRecorder
}

// MockMongoItemMockRecorder is the mock recorder for MockMongoItem.
type MockMongoItemMockRecorder struct {
	mock *MockMongoItem
}

// NewMockMongoItem creates a new mock instance.
func NewMockMongoItem(ctrl *gomock.Controller) *MockMongoItem {
	mock := &MockMongoItem{ctrl: ctrl}
	mock.recorder = &MockMongoItemMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMongoItem) EXPECT() *MockMongoItemMockRecorder {
	return m.recorder
}

// GetCreatedAt mocks base method.
func (m *MockMongoItem) GetCreatedAt() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreatedAt")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetCreatedAt indicates an expected call of GetCreatedAt.
func (mr *MockMongoItemMockRecorder) GetCreatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreatedAt", reflect.TypeOf((*MockMongoItem)(nil).GetCreatedAt))
}

// GetCreatedAtString mocks base method.
func (m *MockMongoItem) GetCreatedAtString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCreatedAtString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetCreatedAtString indicates an expected call of GetCreatedAtString.
func (mr *MockMongoItemMockRecorder) GetCreatedAtString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCreatedAtString", reflect.TypeOf((*MockMongoItem)(nil).GetCreatedAtString))
}

// GetObjectId mocks base method.
func (m *MockMongoItem) GetObjectId() primitive.ObjectID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObjectId")
	ret0, _ := ret[0].(primitive.ObjectID)
	return ret0
}

// GetObjectId indicates an expected call of GetObjectId.
func (mr *MockMongoItemMockRecorder) GetObjectId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObjectId", reflect.TypeOf((*MockMongoItem)(nil).GetObjectId))
}

// GetRandId mocks base method.
func (m *MockMongoItem) GetRandId() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRandId")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRandId indicates an expected call of GetRandId.
func (mr *MockMongoItemMockRecorder) GetRandId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRandId", reflect.TypeOf((*MockMongoItem)(nil).GetRandId))
}

// GetUUID mocks base method.
func (m *MockMongoItem) GetUUID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUUID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetUUID indicates an expected call of GetUUID.
func (mr *MockMongoItemMockRecorder) GetUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUUID", reflect.TypeOf((*MockMongoItem)(nil).GetUUID))
}

// GetUpdatedAt mocks base method.
func (m *MockMongoItem) GetUpdatedAt() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatedAt")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetUpdatedAt indicates an expected call of GetUpdatedAt.
func (mr *MockMongoItemMockRecorder) GetUpdatedAt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatedAt", reflect.TypeOf((*MockMongoItem)(nil).GetUpdatedAt))
}

// GetUpdatedAtString mocks base method.
func (m *MockMongoItem) GetUpdatedAtString() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdatedAtString")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetUpdatedAtString indicates an expected call of GetUpdatedAtString.
func (mr *MockMongoItemMockRecorder) GetUpdatedAtString() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdatedAtString", reflect.TypeOf((*MockMongoItem)(nil).GetUpdatedAtString))
}

// SetCreatedAt mocks base method.
func (m *MockMongoItem) SetCreatedAt(time time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCreatedAt", time)
}

// SetCreatedAt indicates an expected call of SetCreatedAt.
func (mr *MockMongoItemMockRecorder) SetCreatedAt(time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCreatedAt", reflect.TypeOf((*MockMongoItem)(nil).SetCreatedAt), time)
}

// SetCreatedAtString mocks base method.
func (m *MockMongoItem) SetCreatedAtString(timeString string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCreatedAtString", timeString)
}

// SetCreatedAtString indicates an expected call of SetCreatedAtString.
func (mr *MockMongoItemMockRecorder) SetCreatedAtString(timeString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCreatedAtString", reflect.TypeOf((*MockMongoItem)(nil).SetCreatedAtString), timeString)
}

// SetObjectId mocks base method.
func (m *MockMongoItem) SetObjectId() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetObjectId")
}

// SetObjectId indicates an expected call of SetObjectId.
func (mr *MockMongoItemMockRecorder) SetObjectId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetObjectId", reflect.TypeOf((*MockMongoItem)(nil).SetObjectId))
}

// SetRandId mocks base method.
func (m *MockMongoItem) SetRandId() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetRandId")
}

// SetRandId indicates an expected call of SetRandId.
func (mr *MockMongoItemMockRecorder) SetRandId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRandId", reflect.TypeOf((*MockMongoItem)(nil).SetRandId))
}

// SetUUID mocks base method.
func (m *MockMongoItem) SetUUID() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUUID")
}

// SetUUID indicates an expected call of SetUUID.
func (mr *MockMongoItemMockRecorder) SetUUID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUUID", reflect.TypeOf((*MockMongoItem)(nil).SetUUID))
}

// SetUpdatedAt mocks base method.
func (m *MockMongoItem) SetUpdatedAt(time time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdatedAt", time)
}

// SetUpdatedAt indicates an expected call of SetUpdatedAt.
func (mr *MockMongoItemMockRecorder) SetUpdatedAt(time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdatedAt", reflect.TypeOf((*MockMongoItem)(nil).SetUpdatedAt), time)
}

// SetUpdatedAtString mocks base method.
func (m *MockMongoItem) SetUpdatedAtString(timeString string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetUpdatedAtString", timeString)
}

// SetUpdatedAtString indicates an expected call of SetUpdatedAtString.
func (mr *MockMongoItemMockRecorder) SetUpdatedAtString(timeString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUpdatedAtString", reflect.TypeOf((*MockMongoItem)(nil).SetUpdatedAtString), timeString)
}

// MockPagination is a mock of Pagination interface.
type MockPagination[T interfaces.Item] struct {
	ctrl     *gomock.Controller
	recorder *MockPaginationMockRecorder[T]
}

// MockPaginationMockRecorder is the mock recorder for MockPagination.
type MockPaginationMockRecorder[T interfaces.Item] struct {
	mock *MockPagination[T]
}

// NewMockPagination creates a new mock instance.
func NewMockPagination[T interfaces.Item](ctrl *gomock.Controller) *MockPagination[T] {
	mock := &MockPagination[T]{ctrl: ctrl}
	mock.recorder = &MockPaginationMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPagination[T]) EXPECT() *MockPaginationMockRecorder[T] {
	return m.recorder
}

// AddItem mocks base method.
func (m *MockPagination[T]) AddItem(pagKeyParams []string, item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddItem", pagKeyParams, item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// AddItem indicates an expected call of AddItem.
func (mr *MockPaginationMockRecorder[T]) AddItem(pagKeyParams, item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddItem", reflect.TypeOf((*MockPagination[T])(nil).AddItem), pagKeyParams, item)
}

// FetchAll mocks base method.
func (m *MockPagination[T]) FetchAll(pagKeyParams []string, processor interfaces.PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	varargs := []interface{}{pagKeyParams, processor}
	for _, a := range processorArgs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchAll", varargs...)
	ret0, _ := ret[0].([]T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// FetchAll indicates an expected call of FetchAll.
func (mr *MockPaginationMockRecorder[T]) FetchAll(pagKeyParams, processor interface{}, processorArgs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{pagKeyParams, processor}, processorArgs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchAll", reflect.TypeOf((*MockPagination[T])(nil).FetchAll), varargs...)
}

// FetchLinked mocks base method.
func (m *MockPagination[T]) FetchLinked(pagKeyParams, references []string, itemPerPage int64, processor interfaces.PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	varargs := []interface{}{pagKeyParams, references, itemPerPage, processor}
	for _, a := range processorArgs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FetchLinked", varargs...)
	ret0, _ := ret[0].([]T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// FetchLinked indicates an expected call of FetchLinked.
func (mr *MockPaginationMockRecorder[T]) FetchLinked(pagKeyParams, references, itemPerPage, processor interface{}, processorArgs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{pagKeyParams, references, itemPerPage, processor}, processorArgs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchLinked", reflect.TypeOf((*MockPagination[T])(nil).FetchLinked), varargs...)
}

// FetchOne mocks base method.
func (m *MockPagination[T]) FetchOne(randId string) (*T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchOne", randId)
	ret0, _ := ret[0].(*T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// FetchOne indicates an expected call of FetchOne.
func (mr *MockPaginationMockRecorder[T]) FetchOne(randId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchOne", reflect.TypeOf((*MockPagination[T])(nil).FetchOne), randId)
}

// RemoveItem mocks base method.
func (m *MockPagination[T]) RemoveItem(pagKeyParams []string, item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveItem", pagKeyParams, item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// RemoveItem indicates an expected call of RemoveItem.
func (mr *MockPaginationMockRecorder[T]) RemoveItem(pagKeyParams, item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveItem", reflect.TypeOf((*MockPagination[T])(nil).RemoveItem), pagKeyParams, item)
}

// SeedAll mocks base method.
func (m *MockPagination[T]) SeedAll(paginationKeyParameters []string, processor interfaces.PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	varargs := []interface{}{paginationKeyParameters, processor}
	for _, a := range processorArgs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SeedAll", varargs...)
	ret0, _ := ret[0].([]T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// SeedAll indicates an expected call of SeedAll.
func (mr *MockPaginationMockRecorder[T]) SeedAll(paginationKeyParameters, processor interface{}, processorArgs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{paginationKeyParameters, processor}, processorArgs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SeedAll", reflect.TypeOf((*MockPagination[T])(nil).SeedAll), varargs...)
}

// SeedLinked mocks base method.
func (m *MockPagination[T]) SeedLinked(paginationKeyParameters []string, lastItem T, itemPerPage int64, processor interfaces.PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	varargs := []interface{}{paginationKeyParameters, lastItem, itemPerPage, processor}
	for _, a := range processorArgs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SeedLinked", varargs...)
	ret0, _ := ret[0].([]T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// SeedLinked indicates an expected call of SeedLinked.
func (mr *MockPaginationMockRecorder[T]) SeedLinked(paginationKeyParameters, lastItem, itemPerPage, processor interface{}, processorArgs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{paginationKeyParameters, lastItem, itemPerPage, processor}, processorArgs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SeedLinked", reflect.TypeOf((*MockPagination[T])(nil).SeedLinked), varargs...)
}

// SeedOne mocks base method.
func (m *MockPagination[T]) SeedOne(randId string) (*T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SeedOne", randId)
	ret0, _ := ret[0].(*T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// SeedOne indicates an expected call of SeedOne.
func (mr *MockPaginationMockRecorder[T]) SeedOne(randId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SeedOne", reflect.TypeOf((*MockPagination[T])(nil).SeedOne), randId)
}

// TotalItemOnCache mocks base method.
func (m *MockPagination[T]) TotalItemOnCache(pagKeyParams []string) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalItemOnCache", pagKeyParams)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// TotalItemOnCache indicates an expected call of TotalItemOnCache.
func (mr *MockPaginationMockRecorder[T]) TotalItemOnCache(pagKeyParams interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalItemOnCache", reflect.TypeOf((*MockPagination[T])(nil).TotalItemOnCache), pagKeyParams)
}

// UpdateItem mocks base method.
func (m *MockPagination[T]) UpdateItem(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateItem", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// UpdateItem indicates an expected call of UpdateItem.
func (mr *MockPaginationMockRecorder[T]) UpdateItem(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateItem", reflect.TypeOf((*MockPagination[T])(nil).UpdateItem), item)
}

// WithMongo mocks base method.
func (m *MockPagination[T]) WithMongo(mongo interfaces.Mongo[T], paginationFilter bson.A) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WithMongo", mongo, paginationFilter)
}

// WithMongo indicates an expected call of WithMongo.
func (mr *MockPaginationMockRecorder[T]) WithMongo(mongo, paginationFilter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithMongo", reflect.TypeOf((*MockPagination[T])(nil).WithMongo), mongo, paginationFilter)
}

// MockItemCache is a mock of ItemCache interface.
type MockItemCache[T interfaces.Item] struct {
	ctrl     *gomock.Controller
	recorder *MockItemCacheMockRecorder[T]
}

// MockItemCacheMockRecorder is the mock recorder for MockItemCache.
type MockItemCacheMockRecorder[T interfaces.Item] struct {
	mock *MockItemCache[T]
}

// NewMockItemCache creates a new mock instance.
func NewMockItemCache[T interfaces.Item](ctrl *gomock.Controller) *MockItemCache[T] {
	mock := &MockItemCache[T]{ctrl: ctrl}
	mock.recorder = &MockItemCacheMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockItemCache[T]) EXPECT() *MockItemCacheMockRecorder[T] {
	return m.recorder
}

// Del mocks base method.
func (m *MockItemCache[T]) Del(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// Del indicates an expected call of Del.
func (mr *MockItemCacheMockRecorder[T]) Del(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del", reflect.TypeOf((*MockItemCache[T])(nil).Del), item)
}

// Get mocks base method.
func (m *MockItemCache[T]) Get(randId string) (T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", randId)
	ret0, _ := ret[0].(T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockItemCacheMockRecorder[T]) Get(randId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockItemCache[T])(nil).Get), randId)
}

// Set mocks base method.
func (m *MockItemCache[T]) Set(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockItemCacheMockRecorder[T]) Set(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockItemCache[T])(nil).Set), item)
}

// MockMongo is a mock of Mongo interface.
type MockMongo[T interfaces.Item] struct {
	ctrl     *gomock.Controller
	recorder *MockMongoMockRecorder[T]
}

// MockMongoMockRecorder is the mock recorder for MockMongo.
type MockMongoMockRecorder[T interfaces.Item] struct {
	mock *MockMongo[T]
}

// NewMockMongo creates a new mock instance.
func NewMockMongo[T interfaces.Item](ctrl *gomock.Controller) *MockMongo[T] {
	mock := &MockMongo[T]{ctrl: ctrl}
	mock.recorder = &MockMongoMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMongo[T]) EXPECT() *MockMongoMockRecorder[T] {
	return m.recorder
}

// Create mocks base method.
func (m *MockMongo[T]) Create(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockMongoMockRecorder[T]) Create(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockMongo[T])(nil).Create), item)
}

// Delete mocks base method.
func (m *MockMongo[T]) Delete(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockMongoMockRecorder[T]) Delete(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockMongo[T])(nil).Delete), item)
}

// FindMany mocks base method.
func (m *MockMongo[T]) FindMany(filter bson.D, findOptions *options.FindOptions) (*mongo.Cursor, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindMany", filter, findOptions)
	ret0, _ := ret[0].(*mongo.Cursor)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// FindMany indicates an expected call of FindMany.
func (mr *MockMongoMockRecorder[T]) FindMany(filter, findOptions interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindMany", reflect.TypeOf((*MockMongo[T])(nil).FindMany), filter, findOptions)
}

// FindOne mocks base method.
func (m *MockMongo[T]) FindOne(randId string) (T, *commonlogger.CommonError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOne", randId)
	ret0, _ := ret[0].(T)
	ret1, _ := ret[1].(*commonlogger.CommonError)
	return ret0, ret1
}

// FindOne indicates an expected call of FindOne.
func (mr *MockMongoMockRecorder[T]) FindOne(randId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOne", reflect.TypeOf((*MockMongo[T])(nil).FindOne), randId)
}

// GetPaginationFilter mocks base method.
func (m *MockMongo[T]) GetPaginationFilter() bson.A {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPaginationFilter")
	ret0, _ := ret[0].(bson.A)
	return ret0
}

// GetPaginationFilter indicates an expected call of GetPaginationFilter.
func (mr *MockMongoMockRecorder[T]) GetPaginationFilter() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPaginationFilter", reflect.TypeOf((*MockMongo[T])(nil).GetPaginationFilter))
}

// SetPaginationFilter mocks base method.
func (m *MockMongo[T]) SetPaginationFilter(filter bson.A) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPaginationFilter", filter)
}

// SetPaginationFilter indicates an expected call of SetPaginationFilter.
func (mr *MockMongoMockRecorder[T]) SetPaginationFilter(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPaginationFilter", reflect.TypeOf((*MockMongo[T])(nil).SetPaginationFilter), filter)
}

// Update mocks base method.
func (m *MockMongo[T]) Update(item T) *commonlogger.CommonError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", item)
	ret0, _ := ret[0].(*commonlogger.CommonError)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockMongoMockRecorder[T]) Update(item interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockMongo[T])(nil).Update), item)
}
