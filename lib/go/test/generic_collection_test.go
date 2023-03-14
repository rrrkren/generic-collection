package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-emulator/types"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericCollection(t *testing.T) {
	b, keyGen := newTestSetup(t)

	nftKey, nftSigner := keyGen.NewWithSigner()
	nftAddr, metadataAddr, exampleNFTAddr := deployNFTContracts(t, b, []*flow.AccountKey{nftKey}, nftSigner)
	fmt.Printf(
		"contracts deployed: NFT: %s, Metadata: %s, ExampleNFT: %s\n",
		nftAddr,
		metadataAddr,
		exampleNFTAddr,
	)

	genericCollectionAccountKey, genericCollectionSigner := keyGen.NewWithSigner()
	genericCollectionAddr := deployGenericCollectionContract(
		t,
		b,
		[]*flow.AccountKey{genericCollectionAccountKey},
		genericCollectionSigner,
		nftAddr,
	)

	fmt.Printf("generic collection deployed: %s\n", genericCollectionAddr)

	var txResult *types.TransactionResult
	// set exampleNFT in the allowlist
	{
		acct, err := b.GetAccount(genericCollectionAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
	import GenericCollection from 0xGENERICCOLLECTIONADDRESS
	import ExampleNFT from 0xEXAMPLENFTADDRESS
	import NonFungibleToken from 0xNFTADDRESS

	transaction() {
	   prepare(signer: AuthAccount) {
			let admin = signer.borrow<&GenericCollection.Admin>(from: /storage/admin) ?? panic("Could not borrow admin reference")
			admin.setAllowList(t: Type<@ExampleNFT.Collection>(), enabled: true)
	   }
	}
	`
		txCode = strings.ReplaceAll(txCode, "GENERICCOLLECTIONADDRESS", genericCollectionAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "EXAMPLENFTADDRESS", exampleNFTAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "NFTADDRESS", nftAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(genericCollectionAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(genericCollectionAddr).
			AddAuthorizer(genericCollectionAddr)

		txResult = signAndSubmit(t, b, tx, []flow.Address{genericCollectionAddr}, []crypto.Signer{genericCollectionSigner}, false)
	}
	require.NoError(t, txResult.Error)

	// store an example NFT collection into the big collection
	aliceKey, aliceSigner := keyGen.NewWithSigner()
	aliceAddr, err := b.CreateAccount([]*flow.AccountKey{aliceKey}, nil)
	require.NoError(t, err)

	{
		acct, err := b.GetAccount(aliceAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
import GenericCollection from 0xGENERICCOLLECTIONADDRESS
import ExampleNFT from 0xEXAMPLENFTADDRESS
transaction(targetAcct: Address) {
    prepare(signer: AuthAccount) {
		log("hello")
// load the generic collection from target account
		let acct = getAccount(targetAcct)
		let collectionRef = acct.getCapability(/public/genericCollection)!.borrow<&{GenericCollection.PubCollection}>()!

// initialize exampleNFT collection
		let exampleNFTCollection <- ExampleNFT.createEmptyCollection()
		collectionRef.addCollection(collection: <-exampleNFTCollection)
    }
}
`
		txCode = strings.ReplaceAll(txCode, "GENERICCOLLECTIONADDRESS", genericCollectionAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "EXAMPLENFTADDRESS", exampleNFTAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(aliceAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(aliceAddr).
			AddAuthorizer(aliceAddr)

		err = tx.AddArgument(cadence.NewAddress(genericCollectionAddr))
		require.NoError(t, err)

		txResult = signAndSubmit(t, b, tx, []flow.Address{aliceAddr}, []crypto.Signer{aliceSigner}, false)
	}

	require.NoError(t, txResult.Error)
	fmt.Println(txResult.Events)

	// Mint an NFT into the public collection
	{
		acct, err := b.GetAccount(exampleNFTAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
import GenericCollection from 0xGENERICCOLLECTIONADDRESS
import ExampleNFT from 0xEXAMPLENFTADDRESS
import NonFungibleToken from 0xNFTADDRESS
transaction(targetAcct: Address) {
    prepare(signer: AuthAccount) {
		let acct = getAccount(targetAcct)
        let minter = signer.borrow<&ExampleNFT.NFTMinter>(from: ExampleNFT.MinterStoragePath)
            ?? panic("Account does not store an object at the specified path")
		let collectionRef = acct.getCapability(/public/genericCollection)!.borrow<&{GenericCollection.PubCollection}>()!

		let exampleNFTCollection = collectionRef.borrowCollectionPublic(collectionType: Type<@ExampleNFT.Collection>())!
		minter.mintNFT(recipient: exampleNFTCollection, name: "derp", description: "derp", thumbnail: "derp", royalties: [])
    }
}
`
		txCode = strings.ReplaceAll(txCode, "GENERICCOLLECTIONADDRESS", genericCollectionAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "EXAMPLENFTADDRESS", exampleNFTAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "NFTADDRESS", nftAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(exampleNFTAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(exampleNFTAddr).
			AddAuthorizer(exampleNFTAddr)

		err = tx.AddArgument(cadence.NewAddress(genericCollectionAddr))
		require.NoError(t, err)

		txResult = signAndSubmit(t, b, tx, []flow.Address{exampleNFTAddr}, []crypto.Signer{nftSigner}, false)
	}
	require.NoError(t, txResult.Error)
	fmt.Println(txResult.Events)

	// run script to borrow the example NFT
	script := `
import GenericCollection from 0xGENERICCOLLECTIONADDRESS
import ExampleNFT from 0xEXAMPLENFTADDRESS
import NonFungibleToken from 0xNFTADDRESS

/// Script to get NFT IDs in an account's collection
///
pub fun main(): [UInt64] {
    let account = getAccount(0xGENERICCOLLECTIONADDRESS)

    let collectionRef = account
        .getCapability(/public/genericCollection)
        .borrow<&{GenericCollection.PubCollection}>()
        ?? panic("Could not borrow capability from public collection at specified path")
	let exampleNFTCollection = collectionRef.borrowCollectionPublic(collectionType: Type<@ExampleNFT.Collection>())!

    return exampleNFTCollection.getIDs()
}
`
	script = strings.ReplaceAll(script, "GENERICCOLLECTIONADDRESS", genericCollectionAddr.Hex())
	script = strings.ReplaceAll(script, "EXAMPLENFTADDRESS", exampleNFTAddr.Hex())
	script = strings.ReplaceAll(script, "NFTADDRESS", nftAddr.Hex())

	scriptRes, err := b.ExecuteScript([]byte(script), nil)
	require.NoError(t, err)

	nftIDs := scriptRes.Value.ToGoValue().([]interface{})

	require.Len(t, nftIDs, 1)

	exampleNFTID := nftIDs[0].(uint64)

	// withdraw from the generic collection and destroy the NFT
	{
		acct, err := b.GetAccount(genericCollectionAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
	import GenericCollection from 0xGENERICCOLLECTIONADDRESS
	import ExampleNFT from 0xEXAMPLENFTADDRESS
	import NonFungibleToken from 0xNFTADDRESS

	transaction(nftID: UInt64) {
	   prepare(signer: AuthAccount) {
			let GenericCollectionref = signer.borrow<&GenericCollection.Collection>(from: /storage/genericCollection) ?? panic("Could not borrow big collection")

			let exampleNFTCollection = GenericCollectionref.borrowCollection(collectionType: Type<@ExampleNFT.Collection>())! as! &ExampleNFT.Collection
			let nft <- exampleNFTCollection.withdraw(withdrawID: nftID) as! @ExampleNFT.NFT
			log(nft.thumbnail)
			destroy nft
	   }
	}
	`
		txCode = strings.ReplaceAll(txCode, "GENERICCOLLECTIONADDRESS", genericCollectionAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "EXAMPLENFTADDRESS", exampleNFTAddr.Hex())
		txCode = strings.ReplaceAll(txCode, "NFTADDRESS", nftAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(genericCollectionAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(genericCollectionAddr).
			AddAuthorizer(genericCollectionAddr)

		err = tx.AddArgument(cadence.NewUInt64(exampleNFTID))
		require.NoError(t, err)

		txResult = signAndSubmit(t, b, tx, []flow.Address{genericCollectionAddr}, []crypto.Signer{genericCollectionSigner}, false)
	}
	require.NoError(t, txResult.Error)
	fmt.Println(txResult.Events)

}
