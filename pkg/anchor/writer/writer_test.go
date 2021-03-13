/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package writer

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/verifier"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	mockstore "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/sidetree-core-go/pkg/api/operation"
	"github.com/trustbloc/sidetree-core-go/pkg/mocks"

	"github.com/trustbloc/orb/pkg/anchor/graph"
	"github.com/trustbloc/orb/pkg/anchor/txn"
	"github.com/trustbloc/orb/pkg/didtxnref/memdidtxnref"
	vcstore "github.com/trustbloc/orb/pkg/store/verifiable"
)

const (
	namespace = "did:sidetree"

	testDID = "did:method:abc"
)

func TestNew(t *testing.T) {
	var txnCh chan []string

	var vcCh chan *verifiable.Credential

	vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
	require.NoError(t, err)

	providers := &Providers{
		TxnGraph:     graph.New(nil, pubKeyFetcherFnc),
		DidTxns:      memdidtxnref.New(),
		TxnBuilder:   &mockTxnBuilder{},
		ProofHandler: &mockProofHandler{vcCh: make(chan *verifiable.Credential, 100)},
		Store:        vcStore,
	}

	c := New(namespace, providers, txnCh, vcCh)
	require.NotNil(t, c)
}

func TestWriter_WriteAnchor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
		require.NoError(t, err)

		providers := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: make(chan *verifiable.Credential, 100)},
			Store:        vcStore,
		}

		c := New(namespace, providers, txnCh, vcCh)

		err = c.WriteAnchor("anchor", []*operation.Reference{{UniqueSuffix: testDID, Type: operation.TypeCreate}}, 1)
		require.NoError(t, err)

		anchorVC, err := verifiable.ParseCredential([]byte(anchorCred), verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		vcCh <- anchorVC
	})

	t.Run("error - build anchor credential error", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		providersWithErr := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{Err: errors.New("sign error")},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
		}

		c := New(namespace, providersWithErr, txnCh, vcCh)

		err := c.WriteAnchor("anchor", []*operation.Reference{{UniqueSuffix: testDID, Type: operation.TypeCreate}}, 1)
		require.Contains(t, err.Error(), "failed to build anchor credential: sign error")
	})

	t.Run("error - store anchor credential error", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		storeProviderWithErr := mockstore.NewCustomMockStoreProvider(&mockstore.MockStore{
			Store:  make(map[string]mockstore.DBEntry),
			ErrPut: fmt.Errorf("error put"),
		})

		vcStoreWithErr, err := vcstore.New(storeProviderWithErr)
		require.NoError(t, err)

		providersWithErr := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
			Store:        vcStoreWithErr,
		}

		c := New(namespace, providersWithErr, txnCh, vcCh)

		err = c.WriteAnchor("anchor", []*operation.Reference{{UniqueSuffix: testDID, Type: operation.TypeCreate}}, 1)
		require.Contains(t, err.Error(), "error put")
	})

	t.Run("error - previous did transaction reference not found for non-create operations", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
		require.NoError(t, err)

		providers := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
			Store:        vcStore,
		}

		c := New(namespace, providers, txnCh, vcCh)

		err = c.WriteAnchor("anchor", []*operation.Reference{{UniqueSuffix: testDID, Type: operation.TypeUpdate}}, 1)
		require.Contains(t, err.Error(),
			"previous did transaction reference not found for update operation for did[did:method:abc]")
	})
}

func TestWriter_handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
		require.NoError(t, err)

		providers := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: make(chan *verifiable.Credential, 100)},
			Store:        vcStore,
		}

		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		c := New(namespace, providers, txnCh, vcCh)

		anchorVC, err := verifiable.ParseCredential([]byte(anchorCred), verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		c.handle(anchorVC)
	})

	t.Run("error - save anchor credential to store error", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		storeProviderWithErr := mockstore.NewCustomMockStoreProvider(&mockstore.MockStore{
			Store:  make(map[string]mockstore.DBEntry),
			ErrPut: fmt.Errorf("error put"),
		})

		vcStoreWithErr, err := vcstore.New(storeProviderWithErr)
		require.NoError(t, err)

		providersWithErr := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
			Store:        vcStoreWithErr,
		}

		c := New(namespace, providersWithErr, txnCh, vcCh)

		anchorVC, err := verifiable.ParseCredential([]byte(anchorCred), verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		c.handle(anchorVC)
	})

	t.Run("error - add anchor credential to txn graph error", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
		require.NoError(t, err)

		providersWithErr := &Providers{
			TxnGraph:     &mockTxnGraph{Err: errors.New("txn graph error")},
			DidTxns:      memdidtxnref.New(),
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
			Store:        vcStore,
		}

		c := New(namespace, providersWithErr, txnCh, vcCh)

		anchorVC, err := verifiable.ParseCredential([]byte(anchorCred), verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		c.handle(anchorVC)
	})

	t.Run("error - add anchor credential cid to did transactions error", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		vcStore, err := vcstore.New(mockstore.NewMockStoreProvider())
		require.NoError(t, err)

		providersWithErr := &Providers{
			TxnGraph:     graph.New(mocks.NewMockCasClient(nil), pubKeyFetcherFnc),
			DidTxns:      &mockDidTxns{Err: errors.New("did references error")},
			TxnBuilder:   &mockTxnBuilder{},
			ProofHandler: &mockProofHandler{vcCh: vcCh},
			Store:        vcStore,
		}

		c := New(namespace, providersWithErr, txnCh, vcCh)

		anchorVC, err := verifiable.ParseCredential([]byte(anchorCred), verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		c.handle(anchorVC)
	})
}

func TestWriter_Read(t *testing.T) {
	providers := &Providers{
		TxnGraph: graph.New(nil, pubKeyFetcherFnc),
		DidTxns:  memdidtxnref.New(),
	}

	t.Run("success", func(t *testing.T) {
		txnCh := make(chan []string, 100)
		vcCh := make(chan *verifiable.Credential, 100)

		c := New(namespace, providers, txnCh, vcCh)

		more, entries := c.Read(-1)
		require.False(t, more)
		require.Empty(t, entries)
	})
}

type mockTxnBuilder struct {
	Err error
}

func (m *mockTxnBuilder) Build(subject *txn.Payload) (*verifiable.Credential, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &verifiable.Credential{Subject: subject, ID: "http://domain.com/vc/123"}, nil
}

var pubKeyFetcherFnc = func(issuerID, keyID string) (*verifier.PublicKey, error) {
	return nil, nil
}

type mockProofHandler struct {
	vcCh chan *verifiable.Credential
}

func (m *mockProofHandler) RequestProofs(vc *verifiable.Credential, _ []string) error {
	m.vcCh <- vc

	return nil
}

type mockTxnGraph struct {
	Err error
}

func (m *mockTxnGraph) Add(vc *verifiable.Credential) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}

	return "cid", nil
}

type mockDidTxns struct {
	Err error
}

func (m *mockDidTxns) Add(_ []string, _ string) error {
	if m.Err != nil {
		return m.Err
	}

	return nil
}

func (m *mockDidTxns) Last(did string) (string, error) {
	return "cid", nil
}

//nolint:gochecknoglobals,lll
var anchorCred = `
{
	"@context": [
		"https://www.w3.org/2018/credentials/v1"
	],
	"credentialSubject": {
		"anchorString": "1.QmaevShHgc5s7bNnGKkQ98BdaKDNrsCTUV6rcwHr522tQB",
		"namespace": "did:sidetree",
		"previousTransactions": {
		"EiBAnjPBzHqAA-yONCU1HbGln-I0T-ZUPSIkkYAM6EwKKQ": "QmPEVPudBXM5XCoxoNUQiV466e7vD4XowohU8nRAhKJZ6f"
		},
		"version": 0
	},
	"id": "http://peer1.com/vc/85ef42f6-1019-40cc-ab3a-2b477681f5d8",
	"issuanceDate": "2021-03-10T16:34:17.9767297Z",
	"issuer": "http://peer1.com",
	"proof": {
		"created": "2021-03-10T16:34:17.9799878Z",
		"domain": "domain.com",
		"jws": "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..yRt-VlWPDRq0jX-5iMYSfugJspbtmXZn3a9L011w8LI22WzpFZ5YQCTz6B09Stonywg_Xe6fwygG3IPQ5jreBg",
		"proofPurpose": "assertionMethod",
		"type": "Ed25519Signature2018",
		"verificationMethod": "did:web:abc#vaK33R-2ssibOOf2CS0RceLeT61Z2hpskHuEvDW7Hq0"
	},
	"type": "VerifiableCredential"
}`
