/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resthandler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/orb/pkg/activitypub/service/mocks"
	"github.com/trustbloc/orb/pkg/activitypub/store/memstore"
	"github.com/trustbloc/orb/pkg/activitypub/store/spi"
)

const followersURL = "https://example.com/services/orb/followers"

func TestNewFollowers(t *testing.T) {
	serviceIRI, err := url.Parse(serviceURL)
	require.NoError(t, err)

	cfg := &Config{
		BasePath:   basePath,
		ServiceIRI: serviceIRI,
		PageSize:   4,
	}

	h := NewFollowers(cfg, memstore.New(""))
	require.NotNil(t, h)
	require.Equal(t, "/services/orb/followers", h.Path())
	require.Equal(t, http.MethodGet, h.Method())
	require.NotNil(t, h.Handler())
}

func TestFollowers_Handler(t *testing.T) {
	serviceIRI, err := url.Parse(serviceURL)
	require.NoError(t, err)

	followers := newMockURIs(19, func(i int) string { return fmt.Sprintf("https://example%d.com/services/orb", i) })

	activityStore := memstore.New("")

	for _, ref := range followers {
		require.NoError(t, activityStore.AddReference(spi.Follower, serviceIRI, ref))
	}

	cfg := &Config{
		BasePath:   basePath,
		ServiceIRI: serviceIRI,
		PageSize:   4,
	}

	t.Run("Success", func(t *testing.T) {
		h := NewFollowers(cfg, activityStore)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, followersURL, nil)

		h.handle(rw, req)

		result := rw.Result()
		require.Equal(t, http.StatusOK, result.StatusCode)

		respBytes, err := ioutil.ReadAll(result.Body)
		require.NoError(t, err)

		t.Logf("%s", respBytes)

		require.Equal(t, getCanonical(t, followersJSON), getCanonical(t, string(respBytes)))
		require.NoError(t, result.Body.Close())
	})

	t.Run("Store error", func(t *testing.T) {
		cfg := &Config{
			ServiceIRI: serviceIRI,
			PageSize:   4,
		}

		errExpected := fmt.Errorf("injected store error")

		s := &mocks.ActivityStore{}
		s.QueryReferencesReturns(nil, errExpected)

		h := NewFollowers(cfg, s)
		require.NotNil(t, h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, followersURL, nil)

		h.handle(rw, req)

		result := rw.Result()
		require.Equal(t, http.StatusInternalServerError, result.StatusCode)
		require.NoError(t, result.Body.Close())
	})

	t.Run("Marshal error", func(t *testing.T) {
		cfg := &Config{
			ServiceIRI: serviceIRI,
			PageSize:   4,
		}

		h := NewFollowers(cfg, activityStore)
		require.NotNil(t, h)

		errExpected := fmt.Errorf("injected marshal error")

		h.marshal = func(v interface{}) ([]byte, error) {
			return nil, errExpected
		}

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, followersURL, nil)

		h.handle(rw, req)

		result := rw.Result()
		require.Equal(t, http.StatusInternalServerError, result.StatusCode)
		require.NoError(t, result.Body.Close())
	})
}

func TestFollowers_PageHandler(t *testing.T) {
	serviceIRI, err := url.Parse(serviceURL)
	require.NoError(t, err)

	followers := newMockURIs(19, func(i int) string { return fmt.Sprintf("https://example%d.com/services/orb", i+1) })

	activityStore := memstore.New("")

	for _, ref := range followers {
		require.NoError(t, activityStore.AddReference(spi.Follower, serviceIRI, ref))
	}

	t.Run("First page -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "true", "", followersFirstPageJSON)
	})

	t.Run("Page by num -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "true", "3", followersPage3JSON)
	})

	t.Run("Page num too large -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "true", "30", followersPageTooLargeJSON)
	})

	t.Run("Last page -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "true", "4", followersLastPageJSON)
	})

	t.Run("Invalid page-num -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "true", "invalid", followersFirstPageJSON)
	})

	t.Run("Invalid page -> Success", func(t *testing.T) {
		handleFollowersRequest(t, serviceIRI, activityStore, "invalid", "3", followersJSON)
	})

	t.Run("Store error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected store error")

		s := &mocks.ActivityStore{}
		s.QueryReferencesReturns(nil, errExpected)

		cfg := &Config{
			ServiceIRI: serviceIRI,
			PageSize:   4,
		}

		h := NewFollowers(cfg, s)
		require.NotNil(t, h)

		restorePaging := setPaging(h.handler, "true", "0")
		defer restorePaging()

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, followersURL, nil)

		h.handle(rw, req)

		result := rw.Result()
		require.Equal(t, http.StatusInternalServerError, result.StatusCode)
		require.NoError(t, result.Body.Close())
	})

	t.Run("Marshal error", func(t *testing.T) {
		cfg := &Config{
			ServiceIRI: serviceIRI,
			PageSize:   4,
		}

		h := NewFollowers(cfg, activityStore)
		require.NotNil(t, h)

		restorePaging := setPaging(h.handler, "true", "0")
		defer restorePaging()

		errExpected := fmt.Errorf("injected marshal error")

		h.marshal = func(v interface{}) ([]byte, error) {
			return nil, errExpected
		}

		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, followersURL, nil)

		h.handle(rw, req)

		result := rw.Result()
		require.Equal(t, http.StatusInternalServerError, result.StatusCode)
		require.NoError(t, result.Body.Close())
	})
}

func handleFollowersRequest(t *testing.T, serviceIRI *url.URL, as spi.Store, page, pageNum, expected string) {
	cfg := &Config{
		ServiceIRI: serviceIRI,
		PageSize:   4,
	}

	h := NewFollowers(cfg, as)
	require.NotNil(t, h)

	restorePaging := setPaging(h.handler, page, pageNum)
	defer restorePaging()

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, followersURL, nil)

	h.handle(rw, req)

	result := rw.Result()
	require.Equal(t, http.StatusOK, result.StatusCode)

	respBytes, err := ioutil.ReadAll(result.Body)
	require.NoError(t, err)
	require.NoError(t, result.Body.Close())

	t.Logf("%s", respBytes)

	require.Equal(t, getCanonical(t, expected), getCanonical(t, string(respBytes)))
}

const (
	followersJSON = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example1.com/services/orb/followers",
  "type": "Collection",
  "totalItems": 19,
  "first": "https://example1.com/services/orb/followers?page=true",
  "last": "https://example1.com/services/orb/followers?page=true&page-num=4"
}`

	followersFirstPageJSON = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example1.com/services/orb/followers?page=true&page-num=0",
  "type": "CollectionPage",
  "totalItems": 19,
  "next": "https://example1.com/services/orb/followers?page=true&page-num=1",
  "items": [
    "https://example1.com/services/orb",
    "https://example2.com/services/orb",
    "https://example3.com/services/orb",
    "https://example4.com/services/orb"
  ]
}`

	followersLastPageJSON = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example1.com/services/orb/followers?page=true&page-num=4",
  "type": "CollectionPage",
  "totalItems": 19,
  "prev": "https://example1.com/services/orb/followers?page=true&page-num=3",
  "items": [
    "https://example17.com/services/orb",
    "https://example18.com/services/orb",
    "https://example19.com/services/orb"
  ]
}`

	followersPage3JSON = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example1.com/services/orb/followers?page=true&page-num=3",
  "type": "CollectionPage",
  "totalItems": 19,
  "next": "https://example1.com/services/orb/followers?page=true&page-num=4",
  "prev": "https://example1.com/services/orb/followers?page=true&page-num=2",
  "items": [
    "https://example13.com/services/orb",
    "https://example14.com/services/orb",
    "https://example15.com/services/orb",
    "https://example16.com/services/orb"
  ]
}`
	followersPageTooLargeJSON = `{
  "@context": "https://www.w3.org/ns/activitystreams",
  "id": "https://example1.com/services/orb/followers?page=true&page-num=30",
  "type": "CollectionPage",
  "totalItems": 19,
  "prev": "https://example1.com/services/orb/followers?page=true&page-num=4"
}`
)