/*
 * Flow Playground
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package e2eTest

import (
	"fmt"
	"github.com/dapperlabs/flow-playground-api/e2eTest/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContractDeployments(t *testing.T) {
	t.Run("Create deployment for non-existent project", func(t *testing.T) {
		c := newClient()

		badID := uuid.New().String()

		contractA := `
		pub contract HelloWorldA {
			pub var A: String
			pub init() { self.A = "HelloWorldA" }
		}`

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", badID),
			client.Var("script", contractA),
			client.Var("address", addr1),
		)

		assert.Error(t, err)
	})

}

func TestContractTitleParsing(t *testing.T) {
	c := newClient()

	project := createProject(t, c)
	contractA := `
		pub contract HelloWorld {
			pub init() {}
		}`

	var respA CreateContractDeploymentResponse
	err := c.Post(
		MutationCreateContractDeployment,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("script", contractA),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	require.Equal(t, "HelloWorld", respA.CreateContractDeployment.Title)
}

func TestContractRedeployment(t *testing.T) {
	t.Run("same contract name with different arguments", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		contractA := `
		pub contract HelloWorld {
			pub var A: Int
			pub init() { self.A = 5 }
			access(all) fun returnInt(): Int {
        		return self.A
    		}
			access(all) fun setVar(a: Int) {
				self.A = a
			}
		}`

		contractB := `
		pub contract HelloWorld {
			pub var B: String
			pub init() { self.B = "HelloWorldB" }
			access(all) fun returnString(): String {
        		return self.B
    		}
			access(all) fun setVar(b: String) {
				self.B = b
			}
		}`

		var respA CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&respA,
			client.Var("projectId", project.ID),
			client.Var("script", contractA),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, contractA, respA.CreateContractDeployment.Script)

		var respB CreateContractDeploymentResponse
		err = c.Post(
			MutationCreateContractDeployment,
			&respB,
			client.Var("projectId", project.ID),
			client.Var("script", contractB),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, contractB, respB.CreateContractDeployment.Script)

		var accountResp GetAccountResponse
		err = c.Post(
			QueryGetAccount,
			&accountResp,
			client.Var("address", addr1),
			client.Var("projectId", project.ID),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)

		require.Equal(t, []string{"HelloWorld"}, accountResp.Account.DeployedContracts)
	})

	t.Run("Contract redeployment with resource", func(t *testing.T) {
		c := newClient()

		project := createProject(t, c)

		RestNFT := `
		pub contract RestNFT {
			pub let CollectionStoragePath: StoragePath
			pub let CollectionPublicPath: PublicPath
			pub let MinterStoragePath: StoragePath
		
			// Tracks the unique IDs of the NFT
			pub var idCount: UInt64
		
			// Declare the NFT resource type
			pub resource NFT {
				// The unique ID that differentiates each NFT
				pub let id: UInt64
				pub let data: String
		
				// Initialize both fields in the init function
				init(initID: UInt64, str: String) {
					self.id = initID
					self.data = str
				}
		
				pub fun getData(): {String: String} {
					return {"id": self.id.toString(), "data": self.data}
				}
			}
		
			// We define this interface purely as a way to allow users
			// to create public, restricted references to their NFT Collection.
			// They would use this to publicly expose only the deposit, getIDs,
			// and idExists fields in their Collection
			pub resource interface NFTReceiver {
		
				pub fun deposit(token: @NFT)
		
				pub fun getIDs(): [UInt64]
		
				pub fun idExists(id: UInt64): Bool
			}
		
			pub resource Collection: NFTReceiver {
				// dictionary of NFT conforming tokens
				// NFT is a resource type with an UInt64 ID field
				pub var ownedNFTs: @{UInt64: NFT}
		
				// Initialize the NFTs field to an empty collection
				init () {
					self.ownedNFTs <- {}
				}
		
				pub fun withdraw(withdrawID: UInt64): @NFT {
					// If the NFT isn't found, the transaction panics and reverts
					let token <- self.ownedNFTs.remove(key: withdrawID)
						?? panic("Cannot withdraw the specified NFT ID")
		
					return <-token
				}
		
				pub fun deposit(token: @NFT) {
					// add the new token to the dictionary with a force assignment
					// if there is already a value at that key, it will fail and revert
					self.ownedNFTs[token.id] <-! token
				}
		
				// idExists checks to see if a NFT
				// with the given ID exists in the collection
				pub fun idExists(id: UInt64): Bool {
					return self.ownedNFTs[id] != nil
				}
		
				// getIDs returns an array of the IDs that are in the collection
				pub fun getIDs(): [UInt64] {
					return self.ownedNFTs.keys
				}
		
				pub fun getNFTdata(id: UInt64): {String: String} {
					let ref = &self.ownedNFTs[id] as auth &NFT?
					let declared = ref!
					return declared.getData()
				}
		
				pub fun getAll(): [{String: String}] {
					let array: [{String: String}] = []
					for key in self.ownedNFTs.keys {
						let value = self.getNFTdata(id: key)
						array.append(value)
					}
					return array
				}
		
				destroy() {
					destroy self.ownedNFTs
				}
			}
		
			// creates a new empty Collection resource and returns it
			pub fun createEmptyCollection(): @Collection {
				return <- create Collection()
			}
		
			pub fun mintNFT(mintString: String): @NFT {
		
				// create a new NFT
				var newNFT <- create NFT(initID: self.idCount, str: mintString)
		
				self.idCount = self.idCount + 1
		
				return <-newNFT
			}
		
			init() {
				self.CollectionStoragePath = /storage/nftCollection
				self.CollectionPublicPath = /public/nftCollection
				self.MinterStoragePath = /storage/nftMinter
		
				// initialize the ID count to one
				self.idCount = 1
		
				// store an empty NFT Collection in account storage
				self.account.save(<-self.createEmptyCollection(), to: self.CollectionStoragePath)
		
				// publish a reference to the Collection in storage
				self.account.link<&{NFTReceiver}>(self.CollectionPublicPath, target: self.CollectionStoragePath)
			}
		}`

		var resp CreateContractDeploymentResponse
		err := c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", RestNFT),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, RestNFT, resp.CreateContractDeployment.Script)

		err = c.Post(
			MutationCreateContractDeployment,
			&resp,
			client.Var("projectId", project.ID),
			client.Var("script", RestNFT),
			client.Var("address", addr1),
			client.AddCookie(c.SessionCookie()),
		)
		require.NoError(t, err)
		require.Equal(t, RestNFT, resp.CreateContractDeployment.Script)
	})
}

func TestContractInteraction(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	var respA CreateContractDeploymentResponse

	err := c.Post(
		MutationCreateContractDeployment,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("script", counterContract),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	addScript := generateAddTwoToCounterScript(addr1)

	var respB CreateTransactionExecutionResponse

	err = c.Post(
		MutationCreateTransactionExecution,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", addScript),
		client.Var("signers", []string{addr2}),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	assert.Empty(t, respB.CreateTransactionExecution.Errors)
}

func TestContractImport(t *testing.T) {
	c := newClient()

	project := createProject(t, c)

	contractA := `
	pub contract HelloWorldA {
		pub var A: String
		pub init() { self.A = "HelloWorldA" }
	}`

	contractB := `
	import HelloWorldA from 0x01
	pub contract HelloWorldB {
		pub init() {
			log(HelloWorldA.A)
		}
	}`

	var respA CreateContractDeploymentResponse
	err := c.Post(
		MutationCreateContractDeployment,
		&respA,
		client.Var("projectId", project.ID),
		client.Var("script", contractA),
		client.Var("address", addr1),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)

	var respB CreateContractDeploymentResponse
	err = c.Post(
		MutationCreateContractDeployment,
		&respB,
		client.Var("projectId", project.ID),
		client.Var("script", contractB),
		client.Var("address", addr2),
		client.AddCookie(c.SessionCookie()),
	)
	require.NoError(t, err)
	require.Empty(t, respB.CreateContractDeployment.Errors)
}

const counterContract = `
  pub contract Counting {

      pub event CountIncremented(count: Int)

      pub resource Counter {
          pub var count: Int

          init() {
              self.count = 0
          }

          pub fun add(_ count: Int) {
              self.count = self.count + count
              emit CountIncremented(count: self.count)
          }
      }

      pub fun createCounter(): @Counter {
          return <-create Counter()
      }
  }
`

// generateAddTwoToCounterScript generates a script that increments a counter.
// If no counter exists, it is created.
func generateAddTwoToCounterScript(counterAddress string) string {
	return fmt.Sprintf(
		`
            import 0x%s

            transaction {

                prepare(signer: AuthAccount) {
                    if signer.borrow<&Counting.Counter>(from: /storage/counter) == nil {
                        signer.save(<-Counting.createCounter(), to: /storage/counter)
                    }

                    signer.borrow<&Counting.Counter>(from: /storage/counter)!.add(2)
                }
            }
        `,
		counterAddress,
	)
}
