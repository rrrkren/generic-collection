package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onflow/flow-emulator/types"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountController(t *testing.T) {
	b, keyGen := newTestSetup(t)

	nftKey, nftSigner := keyGen.NewWithSigner()
	nftAddr, metadataAddr, exampleNFTAddr := deployNFTContracts(t, b, []*flow.AccountKey{nftKey}, nftSigner)
	fmt.Printf(
		"contracts deployed: NFT: %s, Metadata: %s, ExampleNFT: %s\n",
		nftAddr,
		metadataAddr,
		exampleNFTAddr,
	)

	accountControllerKey, accountControllerSigner := keyGen.NewWithSigner()
	accountControllerAddr := deployAccountControllerContract(
		t,
		b,
		[]*flow.AccountKey{accountControllerKey},
		accountControllerSigner,
		nftAddr,
		exampleNFTAddr,
	)

	fmt.Printf("account controller deployed: %s\n", accountControllerAddr)

	aliceKey, aliceSigner := keyGen.NewWithSigner()
	aliceAddr, err := b.CreateAccount([]*flow.AccountKey{aliceKey}, nil)
	require.NoError(t, err)

	var txResult *types.TransactionResult
	// init example NFT collection
	{
		acct, err := b.GetAccount(aliceAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
	import AccountController from 0xACCOUNTCONTROLLERADDRESS

	transaction() {
	   prepare(signer: AuthAccount) {
		let acct = getAccount(0xACCOUNTCONTROLLERADDRESS)
		let publicController = acct.getCapability(/public/accountController)!.borrow<&{AccountController.PublicController}>()!
		publicController.initExampleToken()
		log("token collection initialized")
	   }
	}
	`
		txCode = strings.ReplaceAll(txCode, "ACCOUNTCONTROLLERADDRESS", accountControllerAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(aliceAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(aliceAddr).
			AddAuthorizer(aliceAddr)

		txResult = signAndSubmit(t, b, tx, []flow.Address{aliceAddr}, []crypto.Signer{aliceSigner}, false)
	}
	require.NoError(t, txResult.Error)
	fmt.Println(txResult.Logs)

	deployAccountControllerUpdatedContract(t, b, []*flow.AccountKey{accountControllerKey}, accountControllerSigner, accountControllerAddr, nftAddr, exampleNFTAddr)

	// init example NFT collection using the new method
	{
		acct, err := b.GetAccount(aliceAddr)
		assert.NoError(t, err)
		acctKey := acct.Keys[0]
		txCode := `
	import AccountController from 0xACCOUNTCONTROLLERADDRESS

	transaction() {
	   prepare(signer: AuthAccount) {
		let acct = getAccount(0xACCOUNTCONTROLLERADDRESS)
		let publicController = acct.getCapability(/public/accountController)!.borrow<&{AccountController.PublicController}>()!
		publicController.initExampleToken2()
		log("token collection initialized")
	   }
	}
	`
		txCode = strings.ReplaceAll(txCode, "ACCOUNTCONTROLLERADDRESS", accountControllerAddr.Hex())

		tx := flow.NewTransaction().
			SetScript([]byte(txCode)).
			SetGasLimit(1000).
			SetProposalKey(aliceAddr, acctKey.Index, acctKey.SequenceNumber).
			SetPayer(aliceAddr).
			AddAuthorizer(aliceAddr)

		txResult = signAndSubmit(t, b, tx, []flow.Address{aliceAddr}, []crypto.Signer{aliceSigner}, false)
	}

}
