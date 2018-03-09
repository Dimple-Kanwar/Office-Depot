/*/*
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

package main

import (
"errors"
"fmt"
"strconv"
"encoding/json"
	//"time"
	//"strings"

"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ManageAccount example simple Chaincode implementation
type ManageAccount struct {
}


var AccountIndexStr = "_AccountIndex"	//name for the key/value that will store a list of all known accounts

type Account struct{
	AccountOwnerId string `json:"accountOwnerId"` 
	AccountName string `json:"accountName"` // Customer or Service Provider
	AccountBalance float64 `json:"accountBalance"`
}
// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(ManageAccount))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *ManageAccount) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error

	// Initialize the chaincode
	
	fmt.Println("ManageAccount chaincode is deployed successfully.")

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(AccountIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	tosend := "{ \"message\" : \"ManageAccount chaincode is deployed successfully.\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	} 
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *ManageAccount) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *ManageAccount) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	}else if function == "createAccount" {											//writes a value to the chaincode state
		return t.createAccount(stub, args)
	}/*else if function == "updateAccountBalance" {									//create a new payment
		return t.updateAccountBalance(stub, args)
	}*/
	fmt.Println("invoke did not find func: " + function)					//error

	errMsg := "{ \"message\" : \"Received unknown function invocation\", \"code\" : \"503\"}"
	err := stub.SetEvent("errEvent", []byte(errMsg))
	if err != nil {
		return nil, err
	} 
	return nil, nil	
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *ManageAccount) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "getAccountByOwner" {													//read a variable
		return t.getAccountByOwner(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	errMsg := "{ \"message\" : \"Received unknown function query\", \"code\" : \"503\"}"
	err := stub.SetEvent("errEvent", []byte(errMsg))
	if err != nil {
		return nil, err
	} 
	return nil, nil
}

// ============================================================================================================================
// Create Account - create a new account for the user, store into chaincode state
// ============================================================================================================================
func (t *ManageAccount) createAccount(stub shim.ChaincodeStubInterface, args []string)([]byte, error){
	
	var err error
	var account Account

	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting \"account details\" as an argument.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		return nil, errors.New(errMsg)
	}
	
	//input sanitation
	if len(args[0]) <= 0 {
		return nil, errors.New("Account details are required")
	}
	accountData := args[0]
	// Converting account details from bytes to Account struct
	json.Unmarshal([]byte(accountData), &account)
	// Fetching account details by account Owner ID
	accountAsBytes, err := stub.GetState(account.AccountOwnerId)
	if err != nil {
		return nil, errors.New("Failed to get Account by Owner ID")
	}
	res := Account{}
	json.Unmarshal(accountAsBytes, &res)
	fmt.Print("Account Details: ")
	fmt.Println(res)
	if res.AccountOwnerId == account.AccountOwnerId{
		fmt.Println("This Account already exists: " + account.AccountOwnerId)
		errMsg := "{ \"message\" : \"This Account already exists.\", \"code\" : \"503\"}"
		err := stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		return nil, errors.New(errMsg)				//stop creating a new account if account exists already
	}

	//build the Account json string manually
	accountJSONasBytes, _ := json.Marshal(account)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	//store AccountOwnerId as key
	err = stub.PutState(account.AccountOwnerId, accountJSONasBytes)	
	if err != nil {
		return nil, err
	}

	//get the Account index
	accountIndexStrAsBytes, err := stub.GetState(AccountIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get Account index")
	}
	var accountIndex []string

	json.Unmarshal(accountIndexStrAsBytes, &accountIndex)							//un stringify it aka JSON.parse()
	fmt.Print("accountIndexStrAsBytes after unmarshal..before append: ")
	fmt.Println(accountIndexStrAsBytes)

	//append
	accountIndex = append(accountIndex, account.AccountOwnerId)									//add AccountOwnerId to index list
	fmt.Println("! Account index: ", accountIndex)
	jsonAsBytes, _ := json.Marshal(accountIndex)
	err = stub.PutState(AccountIndexStr, jsonAsBytes)						//store AccountOwnerId as an index
	if err != nil {
		return nil, err
	}

	tosend := "{ \"Account Owner ID\" : \""+account.AccountOwnerId+"\", \"message\" : \"Account created succcessfully\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	} 

	fmt.Println("Account created succcessfully")
	return nil, nil
}

// ============================================================================================================================
// getAccountByOwner - fetch Account details By Owner Id from chaincode state
// ============================================================================================================================
func (t *ManageAccount) getAccountByOwner(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	fmt.Println("Fetching account by owner Id")
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting \"Account Owner Id\" as an argument.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		return nil, errors.New(errMsg)
	}
	// set accountOwnerId
	accountOwnerId := args[0]
	valAsbytes, err := stub.GetState(accountOwnerId)									//get the accountOwnerId from chaincode state
	if err != nil {
		errMsg := "{ \"message\" : \""+ accountOwnerId + " not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		fmt.Println(errMsg)	
		return nil, errors.New(errMsg)
	}
	fmt.Println("Account details fetched successfully.")
	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// updateAccountBalance - update Account Balance into chaincode state
// ============================================================================================================================
func (t *ManageAccount) updateAccountBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error

	//set amountPaid
	amountPaid := args[2]

	// input sanitation
	if len(args) != 4 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting \"Customer Account Id, Service Provider Account Id, Amount paid\" and \" operation\" as an argument.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		fmt.Println(errMsg)	
		return nil, errors.New(errMsg)
	}

	fmt.Println("Updating the account balance of"+ args[0] + " and " + args[1])
	// convert string to float
	_amountPaid, _ := strconv.ParseFloat(amountPaid, 64)
	operation := arg[3]
	account := Account{}
	for i := 0; i < 2; i++ {
		accountAsBytes, err := stub.GetState(args[i])									//get the var from chaincode state
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + args[i] + "\"}"
			return nil, errors.New(jsonResp)
		}
		json.Unmarshal(accountAsBytes, &account)
		if account.AccountOwnerId == args[i]{
			if account.AccountName == "Customer" {
				fmt.Println("Customer Account found with account Owner Id : " + args[i])
				fmt.Println(account);
				if operation == "Initial" || operation == "Final" {
					account.AccountBalance =  account.AccountBalance - _amountPaid
				}else{
					account.AccountBalance =  account.AccountBalance + _amountPaid
				}
			} else if account.AccountName == "Service Provider" {
				fmt.Println("Service Provider Account found with account Owner Id : " + args[i])
				fmt.Println(account);
				if operation == "Final" || operation == "Initial"{
					account.AccountBalance =  account.AccountBalance + _amountPaid
				}else {
					account.AccountBalance =  account.AccountBalance - _amountPaid
				}
			}
		}else {
			errMsg := "{ \"message\" : \""+ args[i]+ " Not Found.\", \"code\" : \"503\"}"
			err = stub.SetEvent("errEvent", []byte(errMsg))
			if err != nil {
				return nil, err
			}
			fmt.Println(errMsg); 
		}
		
		//build the Payment json string
		accountJson := &Account{account.AccountOwnerId,account.AccountName,account.AccountBalance}
		// convert *Account to []byte
		accountJsonasBytes, err := json.Marshal(accountJson)
		if err != nil {
			return nil, err
		}
		//store account Owner Id as key
		err = stub.PutState(account.AccountOwnerId, accountJsonasBytes)								
		if err != nil {
			return nil, err
		}
		// event message to set on successful account updation
		tosend := "{ \"Account Owner Id\" : \""+account.AccountOwnerId+"\", \"message\" : \"Account updated succcessfully\", \"code\" : \"200\"}"
		err = stub.SetEvent("evtsender", []byte(tosend))
		if err != nil {
			return nil, err
		}
		fmt.Println(tosend); 	
	}
	fmt.Println("Account balance Updated Successfully.")
	return nil, nil
}
