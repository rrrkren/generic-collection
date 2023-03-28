#allowAccountLinking
import NonFungibleToken from "./shared/NonFungibleToken.cdc"
import ExampleNFT from "./shared/ExampleNFT.cdc"

pub contract AccountController {

    pub resource interface PublicController {
      pub fun initExampleToken()
    }

    pub resource Controller: PublicController {
        priv var authAccountCapability: Capability<&AuthAccount>
        init (authAccount: Capability<&AuthAccount>) {
            self.authAccountCapability = authAccount
        }
      pub fun initExampleToken() {
            let account = self.authAccountCapability.borrow()!
            account.save(<-ExampleNFT.createEmptyCollection(), to: /storage/exampleTokenCollection)
            account.link<&{NonFungibleToken.CollectionPublic}>(/public/exampleTokenCollection, target: /storage/exampleTokenCollection)
      }
        destroy() {
        }
    }

    // creates a new empty Collection resource and returns it
    pub fun createController(authAccount: Capability<&AuthAccount>): @Controller {
        return <- create Controller(authAccount: authAccount)
    }

	init() {
        let authAccountCap = self.account.linkAccount(/private/controllerAuthCap)!

        self.account.save(<-self.createController(authAccount: authAccountCap), to: /storage/accountController)
        self.account.link<&{PublicController}>(/public/accountController, target: /storage/accountController)
	}
}
 
