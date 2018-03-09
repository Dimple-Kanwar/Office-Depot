
package main

import (
"errors"
"fmt"
"strconv"
"encoding/json"
"time"
"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ManagePayment example simple Chaincode implementation
type ManagePayment struct {
}

var PaymentIndexStr = "_PaymentIndexStr"

type Payment struct{
	PaymentId string
	agreementId string 
	PaymentType string 
	CustomerAccount string
	ReceiverAccount string
	AmountPaid float64
	LastUpdatedBy string 
	LastUpdateDate int64
}

// ============================================================================================================================
// Main - start the chaincode for Payment management
// ============================================================================================================================
func main() {			
	err := shim.Start(new(ManagePayment))
	if err != nil {
		fmt.Printf("Error starting Payment management chaincode: %s", err)
	}
}
// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *ManagePayment) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(PaymentIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	tosend := "{ \"message\" : \"ManagePayment chaincode is deployed successfully.\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	} 
	fmt.Println("ManagePayment chaincode is deployed successfully.");
	return nil, nil
}
// ============================================================================================================================
// Run - Our entry Paymentint for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *ManagePayment) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}
// ============================================================================================================================
// Invoke - Our entry Paymentint for Invocations
// ============================================================================================================================
func (t *ManagePayment) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	}else if function == "createPayment" {											//create a new  Payment
		return t.createPayment(stub, args)
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
// Query - Our entry Paymentint for Queries
// ============================================================================================================================
func (t *ManagePayment) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "getAll_Payment" {													//Read all  Payments
		return t.getAll_Payment(stub, args)
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
// createPayment - create a new  Payment, store into chaincode state
// ============================================================================================================================
func (t *ManagePayment) createPayment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	if len(args) != 6 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 6 arguments.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		return nil, nil
	}
	fmt.Println("creating a new Payment")
	//input sanitation
	if len(args[0]) <= 0 {
		return nil, errors.New("Agreement Id cannot be empty.")
	}else if len(args[1]) <= 0 {
		return nil, errors.New("Payment Type cannot be empty.")
	}else if len(args[2]) <= 0 {
		return nil, errors.New("Customer Payment cannot be empty.")
	}else if len(args[3]) <= 0 {
		return nil, errors.New("Receiver Payment cannot be empty.")
	}else if len(args[4]) <= 0 {
		return nil, errors.New("Amount Paid cannot be empty.")
	}else if len(args[5]) <= 0 {
		return nil, errors.New("Last Updated By cannot be empty.")
	}

	// setting attributes
	paymentId := "PA"+ strconv.FormatInt(time.Now().Unix(), 10) // check https://play.golang.org/p/8Du2FrDk2eH
	agreementId := args[0]
	paymentType := args[1]
	customerAccount := args[2]
	receiverAccount := args[3]
	amountPaid,_ := strconv.ParseFloat(args[4],64);
	lastUpdatedBy := args[5]
	lastUpdateDate := time.Now().Unix() // current unix timestamp
	
	fmt.Println(paymentId);
	fmt.Println(agreementId);
	fmt.Println(paymentType);
	fmt.Println(customerAccount);
	fmt.Println(receiverAccount);
	fmt.Println(amountPaid);
	fmt.Println(lastUpdatedBy);
	fmt.Println(lastUpdateDate);

	// Fetching Payment details by Payment Id
	PaymentAsBytes, err := stub.GetState(paymentId)
	if err != nil {
		return nil, errors.New("Failed to get Payment Id")
	}
	res := Payment{}
	json.Unmarshal(PaymentAsBytes, &res)
	fmt.Print(" Payment Details: ")
	fmt.Println(res)
	if res.PaymentId == paymentId{
		fmt.Println("This  Payment already exists: " + paymentId)
		errMsg := "{ \"message\" : \"This  Payment already exists.\", \"code\" : \"503\"}"
		err := stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return nil, err
		} 
		return nil, errors.New(errMsg)				//stop creating a new Payment if Payment exists already
	}

	// create a pointer/json to the struct 'Payment'
	PaymentJson := &Payment{paymentId, agreementId, paymentType, customerAccount, receiverAccount, amountPaid, lastUpdatedBy, lastUpdateDate};

	// convert *Payment to []byte
	PaymentJsonasBytes, err := json.Marshal(PaymentJson)
	if err != nil {
		return nil, err
	}
	//store  paymentId as key
	err = stub.PutState(paymentId, PaymentJsonasBytes)								
	if err != nil {
		return nil, err
	}

	//get the Payment index
	PaymentIndexStrAsBytes, err := stub.GetState(PaymentIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get Payment index")
	}
	var PaymentIndex []string

	json.Unmarshal(PaymentIndexStrAsBytes, &PaymentIndex)							//un stringify it aka JSON.parse()
	fmt.Print("PaymentIndex after unmarshal..before append: ")
	fmt.Println(PaymentIndex)

	//append
	PaymentIndex = append(PaymentIndex, paymentId)									//add PaymentId to index list
	fmt.Println("! Payment index: ", PaymentIndex)
	jsonAsBytes, _ := json.Marshal(PaymentIndex)
	err = stub.PutState(PaymentIndexStr, jsonAsBytes)						//store PaymentId as an index
	if err != nil {
		return nil, err
	}

	// event message to set on successful  Payment creation
	tosend := "{ \" Payment Id\" : \""+paymentId+"\", \"message\" : \" Payment created succcessfully\", \"code\" : \"200\"}"
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
	} 	
	fmt.Println(" Payment created succcessfully.")
	return nil, nil
}

// ============================================================================================================================
//  getAll_Payment- get details of all  Payment from chaincode state
// ============================================================================================================================
func (t *ManagePayment) getAll_Payment(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp, errResp string
	var PaymentIndex []string
	fmt.Println("Getting all Payments.")
	var err error
	
	// Fetch all the indexed Payments
	PaymentAsBytes, err := stub.GetState(PaymentIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get Payment index")
	}
	json.Unmarshal(PaymentAsBytes, &PaymentIndex)								//un stringify it aka JSON.parse()
	fmt.Print("PaymentIndex : ")
	fmt.Println(PaymentIndex)
	jsonResp = "{"
	for i,val := range PaymentIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for all Payment")
		valueAsBytes, err := stub.GetState(val)
		if err != nil {
			errResp = "{\"Error\":\"Failed to get state for " + val + "\"}"
			return nil, errors.New(errResp)
		}
		jsonResp = jsonResp + "\""+ val + "\":" + string(valueAsBytes[:])
		if i < len(PaymentIndex)-1 {
			jsonResp = jsonResp + ","
		}
	}
	fmt.Println("len(PaymentIndex) : ")
	fmt.Println(len(PaymentIndex))
	jsonResp = jsonResp + "}"
	fmt.Println("jsonResp : " + jsonResp)
	fmt.Println("Fetched all Payments succcessfully")
	//send it onward
	return []byte(jsonResp), nil
}
