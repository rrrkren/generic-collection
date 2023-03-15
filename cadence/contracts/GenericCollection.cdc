import NonFungibleToken from "./shared/NonFungibleToken.cdc"
pub contract GenericCollection {

  priv let allowList: {Type: Bool}
  pub event CollectionAdded(collectionType: Type)

    pub resource interface PubCollection {
      pub fun addCollection(collection: @AnyResource{NonFungibleToken.CollectionPublic})
      pub fun borrowCollectionPublic(collectionType: Type): &AnyResource{NonFungibleToken.CollectionPublic}?
    }

    // The definition of the Collection resource that
    // holds the NFTs that a user owns
    pub resource Collection: PubCollection {
        // dictionary of NFT conforming tokens
        // NFT is a resource type with an `UInt64` ID field
        pub var collections: @{Type: AnyResource{NonFungibleToken.CollectionPublic}}

        // Initialize the NFTs field to an empty collection
        init () {
            self.collections <- {}
        }

        pub fun withdrawCollection(collectionType: Type): @AnyResource {
            // If the NFT isn't found, the transaction panics and reverts
            let token <- self.collections.remove(key: collectionType)!

            return <-token
        }

        pub fun borrowCollection(collectionType: Type): auth &AnyResource? {
          return &self.collections[collectionType] as auth &AnyResource?
        }


        pub fun borrowCollectionPublic(collectionType: Type): &AnyResource{NonFungibleToken.CollectionPublic}? {
            return &self.collections[collectionType] as &AnyResource{NonFungibleToken.CollectionPublic}?
        }

        pub fun addCollection(collection: @AnyResource{NonFungibleToken.CollectionPublic}) {
            pre {
                GenericCollection.allowList[collection.getType()]! == true: "not in allow list"
            }

            let collectionType = collection.getType()
            let existingCollection <- self.collections.insert(key:collectionType,  <-collection)
            emit CollectionAdded(collectionType: collectionType)
            if existingCollection != nil {
              panic("collection exists")
            }
            destroy existingCollection
        }

        destroy() {
            destroy self.collections
        }
    }

    // creates a new empty Collection resource and returns it 
    pub fun createEmptyCollection(): @Collection {
        return <- create Collection()
    }

    pub resource Admin {
        pub fun setAllowList(t: Type, enabled: Bool) {
            GenericCollection.allowList[t] = enabled
        }
    }

	init() {
        self.allowList = {}
        self.account.save(<-self.createEmptyCollection(), to: /storage/genericCollection)

        // publish a reference to the Collection in storage
        self.account.link<&{PubCollection}>(/public/genericCollection, target: /storage/genericCollection)

        self.account.save(<-create Admin(), to: /storage/admin)
	}
}
 
