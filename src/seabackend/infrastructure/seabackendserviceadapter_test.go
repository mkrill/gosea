package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"github.com/mkrill/gosea/src/seabackend/domain/entity"
	mocks2 "github.com/mkrill/gosea/src/seabackend/infrastructure/mocks"

	"github.com/stretchr/testify/assert"
)

//CacheMock is a manual mocking interface, not necessary when using mockery
//type MockedCache struct{}
//
//func (cm *MockedCache) Get(key string, data interface{}) error {
//	return errors.Errorf("not found")
//}
//
//func (cm *MockedCache) Set(key string, data interface{}) error {
//	return nil
//}

func TestPosts_LoadPosts(t *testing.T) {
	// setup slice tests to contain the test case names and the expected result
	tests := []struct {
		name                   string              // name of test case
		emulatedServerReponse  string              // emulated server response as json string
		emulatedResponseStatus int                 // emulated emulatedResponseStatus
		wrongEndpoint          bool                //
		cacheGetErrorExpected  error               // error expected from mocked get-Cache method
		errorExpected          bool                // true, if error is expected in test case
		expectedResult         []entity.RemotePost // expected RemotePost slice from LoadPosts function
	}{
		{
			name:                   "Normaler Response mit mehreren Werten",
			emulatedServerReponse:  `[{"userId": 1, "id":1, "title": "Title1", "body": "Body1"},{"userId": 2, "id":2, "title": "Title2", "body": "Body2"}]`,
			emulatedResponseStatus: http.StatusOK,
			cacheGetErrorExpected:  errors.New("not found"),
			errorExpected:          false,
			expectedResult: []entity.RemotePost{
				{
					UserID: json.Number("1"),
					ID:     json.Number("1"),
					Title:  "Title1",
					Body:   "Body1",
				},
				{
					UserID: json.Number("2"),
					ID:     json.Number("2"),
					Title:  "Title2",
					Body:   "Body2",
				},
			},
		},
		{
			name:                   "Leerer Response",
			emulatedServerReponse:  ``,
			emulatedResponseStatus: http.StatusOK,
			cacheGetErrorExpected:  errors.New("not found"),
			errorExpected:          true,
			expectedResult:         nil,
		},
		{
			name:                   "Response mit Zahlen als String",
			emulatedServerReponse:  `[{"userId": "1", "id":"1", "title": "Title1", "body": "Body1"},{"userId": 2, "id":2, "title": "Title2", "body": "Body2"}]`,
			emulatedResponseStatus: http.StatusOK,
			cacheGetErrorExpected:  errors.New("not found"),
			errorExpected:          false,
			expectedResult: []entity.RemotePost{
				{
					UserID: json.Number("1"),
					ID:     json.Number("1"),
					Title:  "Title1",
					Body:   "Body1",
				},
				{
					UserID: json.Number("2"),
					ID:     json.Number("2"),
					Title:  "Title2",
					Body:   "Body2",
				},
			},
		},
		{
			name:                   "Falscher Status",
			emulatedServerReponse:  `[{"userId": "1", "id":"1", "title": "Title1", "body": "Body1"},{"userId": 2, "id":2, "title": "Title2", "body": "Body2"}]`,
			emulatedResponseStatus: http.StatusInternalServerError,
			cacheGetErrorExpected:  errors.New("not found"),
			errorExpected:          true,
			expectedResult:         nil,
		},
		{
			name:                  "Falscher Endpunkt",
			wrongEndpoint:         true,
			cacheGetErrorExpected: errors.New("not found"),
			errorExpected:         true,
			expectedResult:        nil,
		},
	}

	// for all test cases in tests slice
	for _, testcase := range tests {
		// run test case testcase, index not used => "_"
		t.Run(testcase.name, func(t *testing.T) {

			// create a new test server with the given function as http handler returning the server
			// response and status assumed in test case
			testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(testcase.emulatedResponseStatus)
				_, err := fmt.Fprint(w, testcase.emulatedServerReponse)
				assert.NoError(t, err)
			}))
			// close the server when finished
			defer testSrv.Close()

			// manual mocking cacher interface alternatively used instead of mocks.Cacher{}
			// mockedCache := MockedCache{}
			mockedCache := mocks2.Cacher{}
			mockedCache.On("Get", mock.Anything, mock.Anything).Return(testcase.cacheGetErrorExpected).Once()
			mockedCache.On("Set", mock.Anything, mock.Anything).Return(nil).Once()

			testPosts := &SeaBackendServiceAdapter{}
			testPosts.Inject(
				&mockedCache,
				&struct {
					SeaEndpoint    string  `inject:"config:seabackend.endpoint"`
					DefaultTimeout float64 `inject:"config:seabackend.defaultTimeout"`
				}{
					SeaEndpoint:    testSrv.URL,
					DefaultTimeout: 100,
				},
			)
			// if test case emulates a wrong endpoint, set it to empty string
			if testcase.wrongEndpoint {
				testPosts.endpoint = ""
			}

			// execute LoadPosts() based on testPosts and check
			// if resulting error is correct
			rp, err := testPosts.LoadPosts(context.TODO())
			if testcase.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// compare expected result based on test case with current result
			assert.Equal(t, testcase.expectedResult, rp)
		})
	}
}
