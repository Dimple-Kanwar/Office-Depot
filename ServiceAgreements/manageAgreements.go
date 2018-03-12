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

package main

import (
"errors"
"fmt"
"strconv"
"encoding/json"
"time"
"github.com/hyperledger/fabric/core/chaincode/shim"
"github.com/hyperledger/fabric/core/util"
)

// ManageAgreement example simple Chaincode implementation
type ManageAgreement struct {
}

var ServiceAgreementIndexStr = "_ServiceAgreementIndexStr"

type Service_agreement struct{
	AgreementID string
	Status string
	CustomerId string
	ServiceProviderId string
	StartDate int64
	EndDate int64
	DueAmount float64
	InitialPaymentPercentage float64
	PenaltyAmount float64
	PenaltyTimePeriod int64
	LastUpdatedBy string
	LastUpdateDate int64
}

type Payment struct{
	PaymentId string
	AgreementId string
	PaymentType string
	CustomerAccount string
	ReceiverAccount string
	AmountPaid float64
	LastUpdatedBy string
	LastUpdateDate int64
}

// ============================================================================================================================
// Main - start the chaincode for Agreement management
// ============================================================================================================================
func main() {
	err := shim.Start(new(ManageAgreement))
	if err != nil {
		fmt.Printf("Error starting Agreement management chaincode: %s", err)
	}
}
// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *ManageAgreement) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting \"Intial_Value\" as an argument\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(ServiceAgreementIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	tosend := "{ \"message\" : \"ManageAgreement chaincode is deployed successfully.\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	}
	fmt.Println("ManageAgreement chaincode is deployed successfully.");
	return nil, nil
}
// ============================================================================================================================
// Run - Our entry agreementint for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *ManageAgreement) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}
// ============================================================================================================================
// Invoke - Our entry agreementint for Invocations
// ============================================================================================================================
func (t *ManageAgreement) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	}else if function == "createServiceAgreement" {											//create a new Service Agreement
		return t.createServiceAgreement(stub, args)
	}else if function == "updateServiceAgreement" {											//update Service Agreement
		return t.updateServiceAgreement(stub, args)
	}else if function == "checkPenalty" {											//update Service Agreement
		return t.checkPenalty(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error
	errMsg := "{ \"message\" : \"Received unknown function invocation\", \"code\" : \"503\"}"
	err := stub.SetEvent("errEvent", []byte(errMsg))
	if err != nil {
		return nil, err
	}
	return nil, nil
}
// ============================================================================================================================
// Query - Our entry agreementint for Queries
// ============================================================================================================================
func (t *ManageAgreement) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "getAll_ServiceAgreement" {													//Read all Service Agreements
		return t.getAll_ServiceAgreement(stub, args)
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
// createServiceAgreement - create a new Service Agreement, store into chaincode state
// ============================================================================================================================
func (t *ManageAgreement) createServiceAgreement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 9 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 9 arguments.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)
	}
	fmt.Println("creating a new Service Agreement")
	//input sanitation
	if len(args[0]) <= 0 {
		return nil, errors.New("Customer Id in an agreement cannot be empty.")
	}else if len(args[1]) <= 0 {
		return nil, errors.New("Service Provider Id in an agreement cannot be empty.")
	}else if len(args[2]) <= 0 {
		return nil, errors.New("Start Date of a Service agreement cannot be empty")
	}else if len(args[3]) <= 0 {
		return nil, errors.New("End Date of a Service agreement cannot be empty.")
	}else if len(args[4]) <= 0 {
		return nil, errors.New("Due Amount of a Service agreement cannot be empty.")
	}else if len(args[5]) <= 0 {
		return nil, errors.New("Initial Payment Percentage cannot be empty.")
	}else if len(args[6]) <= 0 {
		return nil, errors.New("Penalty Amount for a Service agreement cannot be empty.")
	}else if len(args[7]) <= 0 {
		return nil, errors.New("Penalty Time Period of a Service agreement cannot be empty.")
	}else if len(args[8]) <= 0 {
		return nil, errors.New("Last Updated By cannot be empty.")
	}

	// setting attributes
	agreementId := "SA"+ strconv.FormatInt(time.Now().Unix(), 10) // check https://play.golang.org/p/8Du2FrDk2eH
	status := "Pending Customer Acceptance"
	customerId := args[0]
	serviceProviderId := args[1]
	startDate, _ := strconv.ParseInt(args[2], 10,64)
	endDate, _ := strconv.ParseInt(args[3], 10,64)
	dueAmount, _ := strconv.ParseFloat(args[4], 64)
  initialPayment,_ := strconv.ParseFloat(args[5], 64);
	initialPaymentPercentage := initialPayment / 100; // % of the Total amount due
	penaltyAmount, _ := strconv.ParseFloat(args[6], 64)
	penaltyTime,_	:= strconv.ParseFloat(args[7], 64) // minutes in seconds format
	penaltyTimePeriod := int64(penaltyTime)
	lastUpdatedBy := args[8]
	lastUpdateDate := time.Now().Unix() // current unix timestamp

	fmt.Println(agreementId);
	fmt.Println(customerId);
	fmt.Println(serviceProviderId);
	fmt.Println(status);
	fmt.Println(startDate);
	fmt.Println(endDate);
	fmt.Println(dueAmount);
	fmt.Println(initialPaymentPercentage);
	fmt.Println(penaltyAmount);
	fmt.Println(penaltyTimePeriod);
	fmt.Println(lastUpdatedBy);
	fmt.Println(lastUpdateDate);

	// Fetching Service agreement details by agreement Id
	serviceAgreementAsBytes, err := stub.GetState(agreementId)
	if err != nil {
		return nil, errors.New("Failed to get service agreement Id")
	}
	res := Service_agreement{}
	json.Unmarshal(serviceAgreementAsBytes, &res)
	fmt.Print("Service Agreement Details: ")
	fmt.Println(res)
	if res.AgreementID == agreementId{
		fmt.Println("This service agreement already exists: " + agreementId)
		errMsg := "{ \"message\" : \"This service agreement already exists.\", \"code\" : \"503\"}"
		err := stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)				//stop creating a new service agreement if agreement exists already
	}

	// create a pointer/json to the struct 'Service_agreement'
	serviceAgreementJson := &Service_agreement{agreementId, status, customerId, serviceProviderId, startDate, endDate, dueAmount, initialPaymentPercentage, penaltyAmount, penaltyTimePeriod, lastUpdatedBy, lastUpdateDate}

	// convert *Service_agreement to []byte
	serviceAgreementJsonasBytes, err := json.Marshal(serviceAgreementJson)
	if err != nil {
		return nil, err
	}
	//store service agreementId as key
	err = stub.PutState(agreementId, serviceAgreementJsonasBytes)
	if err != nil {
		return nil, err
	}

	//get the Service Agreement index
	serviceAgreementIndexStrAsBytes, err := stub.GetState(ServiceAgreementIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get Service Agreement index")
	}
	var ServiceAgreementIndex []string

	json.Unmarshal(serviceAgreementIndexStrAsBytes, &ServiceAgreementIndex)							//un stringify it aka JSON.parse()
	fmt.Print("serviceAgreementIndexStrAsBytes after unmarshal..before append: ")
	fmt.Println(serviceAgreementIndexStrAsBytes)

	//append
	ServiceAgreementIndex = append(ServiceAgreementIndex, agreementId)									//add agreementId to index list
	fmt.Println("! Service Agreement index: ", ServiceAgreementIndex)
	jsonAsBytes, _ := json.Marshal(ServiceAgreementIndex)
	err = stub.PutState(ServiceAgreementIndexStr, jsonAsBytes)						//store Service Agreement as an index
	if err != nil {
		return nil, err
	}

	// event message to set on successful service agreement creation
	tosend := "{ \"Service Agreement Id\" : \""+agreementId+"\", \"message\" : \"Service agreement created succcessfully\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	}
	fmt.Println("Service agreement created succcessfully.")
	return nil, nil
}

// ============================================================================================================================
// updateServiceAgreement - update Service Agreement into chaincode state
// ============================================================================================================================
func (t *ManageAgreement) updateServiceAgreement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error
	fmt.Println("updating a Service Agreement")
	if len(args) != 5 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 5 arguments.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)
	}
	// set attributes
	agreementId := args[0]
	lastUpdatedBy := args[1]
	newStatus := args[2]
	paymentChaincode := args[3]
	accountChaincode := args[4]
	// Fetch the service agreement details by agreementId
	agreementAsBytes, err := stub.GetState(agreementId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + agreementId + "\"}"
		return nil, errors.New(jsonResp)
	}
	res := Service_agreement{}
	json.Unmarshal(agreementAsBytes, &res)

	if res.AgreementID == agreementId{
		fmt.Println("Agreement found with agreementId : " + agreementId)
		fmt.Println(res);
		res.LastUpdatedBy = lastUpdatedBy
		res.LastUpdateDate = time.Now().Unix() // current unix timestamp
		var paymentStatus string
		// set Payment status according to agreement status
		if res.Status == "Pending Customer Acceptance" && newStatus == "Pending start with Service Provider"{
			paymentStatus = "Initial Payment"
			// Customer account deducted and Service Provider account debited with initial payment
			amountPaid := strconv.FormatFloat(res.DueAmount * res.InitialPaymentPercentage, 'f', 2, 64)
			function := "updateAccountBalance"
			invokeArgs1 := util.ToChaincodeArgs(function, res.CustomerId, res.ServiceProviderId, amountPaid, "Initial")
			update_result, err1 := stub.InvokeChaincode(accountChaincode, invokeArgs1)
			if err1 != nil {
				errStr := fmt.Sprintf("Error in updating account balance from 'Account' chaincode. Got error: %s", err1.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ")
			fmt.Println(update_result);
			fmt.Println("Account Balances updated successfully.");
			fmt.Println(res.AgreementID);
			fmt.Println(agreementId);
			// create Payment transaction
			_function := "createPayment"
			invokeArgs2 := util.ToChaincodeArgs(_function, res.AgreementID, paymentStatus, res.CustomerId, res.ServiceProviderId, amountPaid, lastUpdatedBy)
			result, err2 := stub.InvokeChaincode(paymentChaincode, invokeArgs2)
			if err2 != nil {
				errStr := fmt.Sprintf("Error in fetching Payment details from 'Payment' chaincode. Got error: %s", err2.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ")
			fmt.Println(result);
			fmt.Println("Payment Created successfully.");
		}else if newStatus == "Work in Progress" {
			// no penalty applied
			// do nothing, just update the agreement status
		} else if newStatus == "Work Completed" {
			paymentStatus = "Final Payment"
			//	Customer account deducted with final payment (total amount – initial payment)
			//	Service Provider account credited with final payment
			amountPaid := strconv.FormatFloat(res.DueAmount -(res.DueAmount * res.InitialPaymentPercentage), 'f', 2, 64)
			function := "updateAccountBalance"
			invokeArgs1 := util.ToChaincodeArgs(function, res.CustomerId, res.ServiceProviderId, amountPaid, "Final")
			update_result, err1 := stub.InvokeChaincode(accountChaincode, invokeArgs1)
			if err1 != nil {
				errStr := fmt.Sprintf("Error in updating account balance from 'Account' chaincode. Got error: %s", err1.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ")
			fmt.Println(update_result);
			fmt.Println("Account Balances updated successfully.");
			// create Payment transaction
			_function := "createPayment"
			invokeArgs2 := util.ToChaincodeArgs(_function, res.AgreementID, paymentStatus, res.CustomerId, res.ServiceProviderId, amountPaid, lastUpdatedBy)
			result, err2 := stub.InvokeChaincode(paymentChaincode, invokeArgs2)
			if err2 != nil {
				errStr := fmt.Sprintf("Error in fetching Payment details from 'Payment' chaincode. Got error: %s", err2.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ")
			fmt.Println(result);
			fmt.Println("Payment Created successfully.");
		}
		//build the Service Agreement json
		serviceAgreementJson := &Service_agreement{res.AgreementID, newStatus, res.CustomerId, res.ServiceProviderId, res.StartDate, res.EndDate, res.DueAmount, res.InitialPaymentPercentage, res.PenaltyAmount, res.PenaltyTimePeriod, res.LastUpdatedBy, res.LastUpdateDate}

		// convert *Service_agreement to []byte
		serviceAgreementJsonasBytes, err := json.Marshal(serviceAgreementJson)
		if err != nil {
			return nil, err
		}
		//store Agreement id as key
		err = stub.PutState(res.AgreementID, serviceAgreementJsonasBytes)
		if err != nil {
			return nil, err
		}
		tosend := "{ \"Service Agreement ID\" : \""+res.AgreementID+"\", \"message\" : \"Service Agreement updated succcessfully\", \"code\" : \"200\"}"
		err = stub.SetEvent("evtsender", []byte(tosend))
		if err != nil {
			return nil, err
		}
	}else{
		errMsg := "{ \"message\" : \""+ agreementId+ " Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	fmt.Println("updated Service Agreement")
	return nil, nil
}

// ============================================================================================================================
// CheckPenalty - update Service Agreement into chaincode state
// ============================================================================================================================
func (t *ManageAgreement) checkPenalty(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error
	fmt.Println("Penalty Check Started.")
	if len(args) != 4 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 4 arguments.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errMsg)
	}
	// set attributes
	agreementId := args[0]
	lastUpdatedBy := args[1]
	paymentChaincode := args[2]
	accountChaincode := args[3]

	// Fetch the service agreement details by agreementId
	agreementAsBytes, err := stub.GetState(agreementId)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + agreementId + "\"}"
		return nil, errors.New(jsonResp)
	}
	res := Service_agreement{}
	json.Unmarshal(agreementAsBytes, &res)

	if res.AgreementID == agreementId{
		fmt.Println("Agreement found with agreementId : " + agreementId)
		fmt.Println(res);
		/*currentTime := time.Now().Unix()
		fmt.Println(currentTime);
		fmt.Println(res.LastUpdateDate);
		fmt.Println(currentTime - res.LastUpdateDate);*/
		if res.Status == "Pending start with Service Provider" /*&& res.PenaltyTimePeriod < currentTime - res.LastUpdateDate*/{
			//	Service Provider account deducted with penalty amount
			amountPaid := res.PenaltyAmount
			paymentStatus := "Penalty Payment"
			function := "updateAccountBalance"
			invokeArgs := util.ToChaincodeArgs(function, res.CustomerId, res.ServiceProviderId, strconv.FormatFloat(res.PenaltyAmount,'f', 2, 64),"Penalty")
			update_result, err := stub.InvokeChaincode(accountChaincode, invokeArgs)
			if err != nil {
				errStr := fmt.Sprintf("Error in updating account balance from 'Account' chaincode. Got error: %s", err.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ",update_result);
			fmt.Println("Account Balances updated successfully.");
			// create Payment transaction
			_function := "createPayment"
			invokeArgs1 := util.ToChaincodeArgs(_function, res.AgreementID, paymentStatus, res.CustomerId, res.ServiceProviderId,  strconv.FormatFloat(amountPaid,'f', 2, 64), lastUpdatedBy)
			result, err1 := stub.InvokeChaincode(paymentChaincode, invokeArgs1)
			if err1 != nil {
				errStr := fmt.Sprintf("Error in fetching Payment details from 'Payment' chaincode. Got error: %s", err1.Error())
				fmt.Printf(errStr)
				return nil, errors.New(errStr)
			}
			fmt.Println("transaction Hash: ", result);
			fmt.Println("Penalty Payment Created successfully.");
			tosend := "{ \"Service Agreement Id\" : \""+agreementId+"\", \"message\" : \"Penalty Applied to the agreement.\", \"code\" : \"200\"}"
			err = stub.SetEvent("evtsender", []byte(tosend))
			if err != nil {
				return nil, err
			}
			fmt.Println(tosend);
		}else{
			tosend := "{ \"Service Agreement Id\" : \""+agreementId+"\", \"message\" : \"Penalty cannot be applied to the agreement.\", \"code\" : \"200\"}"
			err = stub.SetEvent("evtsender", []byte(tosend))
			if err != nil {
				return nil, err
			}
			fmt.Println(tosend);
		}
	}else{
		errMsg := "{ \"message\" : \""+ agreementId+ " Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	fmt.Println("Penalty Check Completed.");
	return nil, nil
}
// ============================================================================================================================
//  getAll_ServiceAgreement- get details of all Service Agreement from chaincode state
// ============================================================================================================================
func (t *ManageAgreement) getAll_ServiceAgreement(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp, errResp string
	var agreementIndex []string
	fmt.Println("start getAll_ServiceAgreement")
	var err error
	// if len(args) != 1 {
	// 	errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting \" \" as an argument\", \"code\" : \"503\"}"
	// 	err = stub.SetEvent("errEvent", []byte(errMsg))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return nil, nil
	// }

	// Fetch all the indexed Service agreements
	agreementAsBytes, err := stub.GetState(ServiceAgreementIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get Agreement index")
	}
	json.Unmarshal(agreementAsBytes, &agreementIndex)								//un stringify it aka JSON.parse()
	fmt.Print("agreementIndex : ")
	fmt.Println(agreementIndex)
	jsonResp = "{"
	for i,val := range agreementIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for all Agreement")
		valueAsBytes, err := stub.GetState(val)
		if err != nil {
			errResp = "{\"Error\":\"Failed to get state for " + val + "\"}"
			return nil, errors.New(errResp)
		}
		jsonResp = jsonResp + "\""+ val + "\":" + string(valueAsBytes[:])
		if i < len(agreementIndex)-1 {
			jsonResp = jsonResp + ","
		}
	}
	fmt.Println("len(agreementIndex) : ")
	fmt.Println(len(agreementIndex))
	jsonResp = jsonResp + "}"
	fmt.Println("jsonResp : " + jsonResp)
	fmt.Println("end get_AllAgreement")
	//send it onward
	return []byte(jsonResp), nil
}
