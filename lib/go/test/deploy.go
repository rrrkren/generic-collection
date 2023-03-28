package test

import (
	"rrrkren/generic-collection/lib/go/contracts"
	"testing"

	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	nftcontracts "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/stretchr/testify/require"
)

func deployGenericCollectionContract(
	t *testing.T,
	b *emulator.Blockchain,
	key []*flow.AccountKey,
	signer crypto.Signer,
	nftAddr flow.Address,
) flow.Address {

	contractCode := contracts.GenericCollection(nftAddr.String())
	contractAddress, err := b.CreateAccount(key, []sdktemplates.Contract{
		{
			Name:   "GenericCollection",
			Source: string(contractCode),
		},
	})
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return contractAddress
}

func deployAccountControllerContract(
	t *testing.T,
	b *emulator.Blockchain,
	key []*flow.AccountKey,
	signer crypto.Signer,
	nftAddr flow.Address,
	exampleTokenAddr flow.Address,
) flow.Address {

	contractCode := contracts.AccountController(nftAddr.String(), exampleTokenAddr.String())
	contractAddress, err := b.CreateAccount(key, []sdktemplates.Contract{
		{
			Name:   "AccountController",
			Source: string(contractCode),
		},
	})
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return contractAddress
}

func deployNFTContracts(
	t *testing.T,
	b *emulator.Blockchain,
	key []*flow.AccountKey,
	signer crypto.Signer,
) (
	flow.Address,
	flow.Address,
	flow.Address,
) {
	nftCode := nftcontracts.NonFungibleToken()
	nftAddress, err := b.CreateAccount(key,
		[]sdktemplates.Contract{
			{
				Name:   "NonFungibleToken",
				Source: string(nftCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	metadataCode := nftcontracts.MetadataViews(flow.HexToAddress(contracts.EmulatorFungibleTokenAddress), nftAddress)

	metadataAddr, err := b.CreateAccount(key,
		[]sdktemplates.Contract{
			{
				Name:   "MetadataViews",
				Source: string(metadataCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	exampleNFTCode := nftcontracts.ExampleNFT(nftAddress, metadataAddr)
	exampleNFTAddress, err := b.CreateAccount(key,
		[]sdktemplates.Contract{
			{
				Name:   "ExampleNFT",
				Source: string(exampleNFTCode),
			},
		},
	)
	require.NoError(t, err)

	_, err = b.CommitBlock()
	require.NoError(t, err)

	return nftAddress, metadataAddr, exampleNFTAddress
}
