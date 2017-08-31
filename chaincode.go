/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

//The following examples demonstrate creating indexes on CouchDB
//Example hostname:port configurations
//
//Docker or vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for chaincodeid, docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":["chaincodeid","data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1/_index
//

// Index for chaincodeid, docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"chaincodeid":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.size\":\"desc\"},{\"chaincodeid\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// 매물
type property struct {
	ObjectType         string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Property_num       int `json:"property_num"`    //the fieldtags are needed to keep case from bouncing around
	Name							 string `json:"name"`
	Address            string `json:"address"`
	Owner              string    `json:"owner"`
}

// 계약 조건
type conditionOfContract struct {
	ObjectType         string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Condition_num       int `json:"condition_num"`    //the fieldtags are needed to keep case from bouncing around
	Property_num       int `json:"property_num"`
	Seller             string `json:"seller"`
  Buyer              string `json:"buyer"`
  Deposit            int `json:"deposit"`
}

// 계약서
type contract struct {
	ObjectType         string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Contract_num       int `json:"contract_num"`    //the fieldtags are needed to keep case from bouncing around
	Condition_num      int `json:"condition_num"`
}


// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initMarble" {
		return t.initMarble(stub, args)
	} else if function == "transferMarble" { //change owner of a specific property
		return t.transferMarble(stub, args)
	} else if function == "readMarble" {
		return t.readMarble(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initContract
// ============================================================
func (t *SimpleChaincode) initContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// property
	propertyNum := 1
	propertyName := "ISLAB"
	address := "서울시 성북구"
	owner := "A"

	// conditon of contract
	conditionNum := 1
		  // propertyNum
	seller := "A"
	buyer := "B"
	deposit := 5000000

	// contract
	contractNum := 1
		// conditionNum

	// ==== Create property, condition of contract, contract object and marshal to JSON ====
	// property object
	objectType := "property"
	property := &property{objectType, propertyNum, propertyName, address, owner}
	propertyJSONasBytes, err := json.Marshal(property)
	if err != nil {
		return shim.Error(err.Error())

	// condition of contract object
	objectType = "conditionOfContract"
	condition := &conditionOfContract{objectType, conditionNum, propertyNum, seller, buyer, deposit}
	conditionJSONasBytes, err := json.Marshal(condition)
	if err != nil {
		return shim.Error(err.Error())

	// contract object
	objectType = "contract"
	contract := &contract{objectType, contractNum, conditionNum}
	contractJSONasBytes, err := json.Marshal(contract)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save object to state ===
	err = stub.PutState(propertyNum, propertyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(conditionNum, conditionOfContractJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(contractNum, contractJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Return success ====
	fmt.Println("- end init contract")
	return shim.Success(nil)
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (t *SimpleChaincode) readProperty(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var num, jsonResp string
	var err error

	num = strconv.Atoi(args[0])

	// === Create composit key ===
	propertyCompositeKey, err := stub.CreateCompositeKey("property", "Property_num")
	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if propertyAsBytes == nil {
		return shim.Error("Property does not exist")
	}

	valAsbytes, err := stub.GetStateByPartialCompositeKey(propertyCompositeKey)
	/*
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}
*/
	return shim.Success(valAsbytes)
}

// ===========================================================
// transfer a property by setting a new owner name on the property
// ===========================================================
func (t *SimpleChaincode) transferProperty(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	propertyNum := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarble ", propertyNum, newOwner)

	propertyCompositeKey, err := stub.CreateCompositeKey("property", "Property_num")
	if err != nil {
		fmt.Println(err.Error())
	}

	propertyAsBytes, err := stub.GetStateByPartialCompositeKey(propertyCompositeKey)

	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if propertyAsBytes == nil {
		return shim.Error("Property does not exist")
	}

	propertyToTransfer := property{}
	err = json.Unmarshal(propertyAsBytes, &propertyToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	propertyToTransfer.Owner = newOwner //change the owner

	propertyJSONasBytes, _ := json.Marshal(propertyToTransfer)
	err = stub.PutState(propertyNum, propertyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferProperty (success)")
	return shim.Success(nil)
}
