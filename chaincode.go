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
	if function == "initProperty" {
		return t.initProperty(stub, args)
	} else if function == "initConditon" {
		return t.initConditon(stub, args)
	} else if function == "CreateContract" {
		return t.CreateContract(stub, args)
	} else if function == "transferProperty" {
		return t.transferProperty(stub, args)
	} else if function == "readValue" {
		return t.readValue(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initProperty
// ============================================================
func (t *SimpleChaincode) initProperty(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// propertyNum, propertyName, address, owner
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	// property
	propertyNum := strings.ToLower(args[0])
	propertyName := strings.ToLower(args[1])
	address := strings.ToLower(args[2])
	owner := strings.ToLower(args[3])

	// ==== Create property object and marshal to JSON ====
	objectType := "property"
	property := &property{objectType, propertyNum, propertyName, address, owner}
	propertyJSONasBytes, err := json.Marshal(property)
	if err != nil {
		return shim.Error(err.Error())
	}
	// === Save object to state ===
	err = stub.PutState(propertyNum, propertyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Return success ====
	fmt.Println("- end init Property")
	return shim.Success(nil)
}

// ============================================================
// initConditon
// ============================================================
func (t *SimpleChaincode) initConditon(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// conditionNum, propertyNum, seller, buyer, deposit
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init condition")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}

	// condition
	conditionNum := strings.ToLower(args[0])
	propertyNum := strings.ToLower(args[1])
	seller := strings.ToLower(args[2])
	buyer := strings.ToLower(args[3])
	deposit, err :=strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("5th argument must be a numeric string")
	}

	// ==== Create condition object and marshal to JSON ====
	objectType := "condition"
	condition := &conditionOfContract{objectType, conditionNum, propertyNum, seller, buyer, deposit}
	conditionJSONasBytes, err := json.Marshal(condition)
	if err != nil {
		return shim.Error(err.Error())

	// === Save object to state ===
	err = stub.PutState(conditionNum, conditionJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Return success ====
	fmt.Println("- end init contract condition")
	return shim.Success(nil)
}

// ============================================================
// CreateContract
// ============================================================
func (t *SimpleChaincode) CreateContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// contractNum, conditionNum
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init condition")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	// contract
	contractNum := strings.ToLower(args[0])
	propertyNum := strings.ToLower(args[1])

	// ==== Create contract object and marshal to JSON ====
	objectType := "contract"
	contract := &conditionOfContract{objectType, conditionNum, propertyNum, seller, buyer, deposit}
	contractJSONasBytes, err := json.Marshal(contract)
	if err != nil {
		return shim.Error(err.Error())

	// === Save object to state ===
	err = stub.PutState(contractNum, contractJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Return success ====
	fmt.Println("- end create contract")
	return shim.Success(nil)
}

// ===============================================
// readValue - read a property, condition, contract from chaincode state
// ===============================================
func (t *SimpleChaincode) readValue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting number of the value to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get value for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Value does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ===========================================================
// transfer a property by setting a new owner name on the property
// ===========================================================
func (t *SimpleChaincode) transferProperty(stub shim.ChaincodeStubInterface, args []string) pb.Response {

		//   0       1
		// "name", "bob"
		if len(args) < 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2")
		}

		propertyNum := args[0]
		newOwner := strings.ToLower(args[1])
		fmt.Println("- start transferProperty ", propertyNum, newOwner)

		propertyAsBytes, err := stub.GetState(propertyNum)
		if err != nil {
			return shim.Error("Failed to get property:" + err.Error())
		} else if propertyAsBytes == nil {
			return shim.Error("Property does not exist")
		}

		propertyToTransfer := marble{}
		err = json.Unmarshal(propertyAsBytes, &propertyToTransfer) //unmarshal it aka JSON.parse()
		if err != nil {
			return shim.Error(err.Error())
		}
		propertyToTransfer.Owner = newOwner //change the owner

		propertyJSONasBytes, _ := json.Marshal(propertyToTransfer)
		err = stub.PutState(propertyNum, propertyJSONasBytes) //rewrite the property
		if err != nil {
			return shim.Error(err.Error())
		}

		fmt.Println("- end transferProperty (success)")
		return shim.Success(nil)
}
