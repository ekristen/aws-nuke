package resources

import (
	"context"
	"reflect"

	"go.uber.org/mock/gomock"
)

// MockTransformCustomAPI is a mock of TransformCustomAPI interface.
type MockTransformCustomAPI struct {
	ctrl     *gomock.Controller
	recorder *MockTransformCustomAPIMockRecorder
}

// MockTransformCustomAPIMockRecorder is the mock recorder for MockTransformCustomAPI.
type MockTransformCustomAPIMockRecorder struct {
	mock *MockTransformCustomAPI
}

// NewMockTransformCustomAPI creates a new mock instance.
func NewMockTransformCustomAPI(ctrl *gomock.Controller) *MockTransformCustomAPI {
	mock := &MockTransformCustomAPI{ctrl: ctrl}
	mock.recorder = &MockTransformCustomAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTransformCustomAPI) EXPECT() *MockTransformCustomAPIMockRecorder {
	return m.recorder
}

// ListCampaigns mocks base method.
func (m *MockTransformCustomAPI) ListCampaigns(
	ctx context.Context, params *TransformCustomListCampaignsInput,
) (*TransformCustomListCampaignsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCampaigns", ctx, params)
	ret0, _ := ret[0].(*TransformCustomListCampaignsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListCampaigns indicates an expected call of ListCampaigns.
func (mr *MockTransformCustomAPIMockRecorder) ListCampaigns(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "ListCampaigns",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).ListCampaigns), ctx, params)
}

// DeleteCampaign mocks base method.
func (m *MockTransformCustomAPI) DeleteCampaign(
	ctx context.Context, params *TransformCustomDeleteCampaignInput,
) (*TransformCustomDeleteCampaignOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteCampaign", ctx, params)
	ret0, _ := ret[0].(*TransformCustomDeleteCampaignOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteCampaign indicates an expected call of DeleteCampaign.
func (mr *MockTransformCustomAPIMockRecorder) DeleteCampaign(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "DeleteCampaign",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).DeleteCampaign), ctx, params)
}

// ListTransformationPackageMetadata mocks base method.
func (m *MockTransformCustomAPI) ListTransformationPackageMetadata(
	ctx context.Context, params *TransformCustomListTransformationPackageMetadataInput,
) (*TransformCustomListTransformationPackageMetadataOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTransformationPackageMetadata", ctx, params)
	ret0, _ := ret[0].(*TransformCustomListTransformationPackageMetadataOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTransformationPackageMetadata indicates an expected call of ListTransformationPackageMetadata.
func (mr *MockTransformCustomAPIMockRecorder) ListTransformationPackageMetadata(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "ListTransformationPackageMetadata",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).ListTransformationPackageMetadata), ctx, params)
}

// DeleteTransformationPackage mocks base method.
func (m *MockTransformCustomAPI) DeleteTransformationPackage(
	ctx context.Context, params *TransformCustomDeleteTransformationPackageInput,
) (*TransformCustomDeleteTransformationPackageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTransformationPackage", ctx, params)
	ret0, _ := ret[0].(*TransformCustomDeleteTransformationPackageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteTransformationPackage indicates an expected call of DeleteTransformationPackage.
func (mr *MockTransformCustomAPIMockRecorder) DeleteTransformationPackage(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "DeleteTransformationPackage",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).DeleteTransformationPackage), ctx, params)
}

// ListKnowledgeItems mocks base method.
func (m *MockTransformCustomAPI) ListKnowledgeItems(
	ctx context.Context, params *TransformCustomListKnowledgeItemsInput,
) (*TransformCustomListKnowledgeItemsOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListKnowledgeItems", ctx, params)
	ret0, _ := ret[0].(*TransformCustomListKnowledgeItemsOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListKnowledgeItems indicates an expected call of ListKnowledgeItems.
func (mr *MockTransformCustomAPIMockRecorder) ListKnowledgeItems(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "ListKnowledgeItems",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).ListKnowledgeItems), ctx, params)
}

// DeleteKnowledgeItem mocks base method.
func (m *MockTransformCustomAPI) DeleteKnowledgeItem(
	ctx context.Context, params *TransformCustomDeleteKnowledgeItemInput,
) (*TransformCustomDeleteKnowledgeItemOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteKnowledgeItem", ctx, params)
	ret0, _ := ret[0].(*TransformCustomDeleteKnowledgeItemOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteKnowledgeItem indicates an expected call of DeleteKnowledgeItem.
func (mr *MockTransformCustomAPIMockRecorder) DeleteKnowledgeItem(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "DeleteKnowledgeItem",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).DeleteKnowledgeItem), ctx, params)
}

// ListCampaignRepositories mocks base method.
func (m *MockTransformCustomAPI) ListCampaignRepositories(
	ctx context.Context, params *TransformCustomListCampaignRepositoriesInput,
) (*TransformCustomListCampaignRepositoriesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListCampaignRepositories", ctx, params)
	ret0, _ := ret[0].(*TransformCustomListCampaignRepositoriesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListCampaignRepositories indicates an expected call of ListCampaignRepositories.
func (mr *MockTransformCustomAPIMockRecorder) ListCampaignRepositories(ctx, params any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock, "ListCampaignRepositories",
		reflect.TypeOf((*MockTransformCustomAPI)(nil).ListCampaignRepositories), ctx, params)
}
