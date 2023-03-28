package contracts

import (
	"fmt"
	"regexp"
	"rrrkren/generic-collection/lib/go/contracts/internal/assets"
)

//go:generate go run github.com/kevinburke/go-bindata/go-bindata -prefix ../../../cadence/contracts -o internal/assets/assets.go -pkg assets -nometadata -nomemcopy ../../../cadence/contracts/...

const (
	filenameGenericCollection = "GenericCollection.cdc"
	filenameAccountController = "AccountController.cdc"

	EmulatorFlowTokenAddress        = "0x0ae53cb6e3f42a79"
	EmulatorFungibleTokenAddress    = "0xee82856bf20e2aa6"
	EmulatorNonFungibleTokenAddress = "0xf8d6e0586b0a20c7"
)

var (
	placeHolderNonFungibleToken = regexp.MustCompile(`".*/NonFungibleToken.cdc"`)
	placeHolderExampleToken     = regexp.MustCompile(`".*/ExampleNFT.cdc"`)
)

func withHexPrefix(address string) string {
	if address == "" {
		return ""
	}

	if address[0:2] == "0x" {
		return address
	}

	return fmt.Sprintf("0x%s", address)
}

func GenericCollection(nftAddr string) []byte {
	code := assets.MustAssetString(filenameGenericCollection)

	code = placeHolderNonFungibleToken.ReplaceAllString(code, withHexPrefix(nftAddr))

	return []byte(code)
}

func AccountController(nftAddr, exampleTokenAddr string) []byte {
	code := assets.MustAssetString(filenameAccountController)
	code = placeHolderNonFungibleToken.ReplaceAllString(code, withHexPrefix(nftAddr))
	code = placeHolderExampleToken.ReplaceAllString(code, withHexPrefix(exampleTokenAddr))

	return []byte(code)
}
