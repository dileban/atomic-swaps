package main

import (
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	tokens "github.com/dileban/atomic-swaps/fabric/lib/asset/fungible"
	"github.com/dileban/atomic-swaps/fabric/lib/security"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// TokenChaincode is ... implements shim.Chaincode
type TokenChaincode struct {
	token tokens.SimpleToken
}

// CallerProps is a container for meta data from the remote client as
// well as the peer. This includes the arguments and identity of the
// client as well as callback pointers to the peer.
type CallerProps struct {
	args []string
	cert *x509.Certificate
	stub shim.ChaincodeStubInterface
}

// initialOwner is the address of the initial owner of the token
// supply. If specified, the Init function checks to see of the
// supplied owner address matches. Its value must be specified before
// the multi-org chaincode package signing process begins.
const initialOwner = ""

// For use within handlers and the token implementation.
var caller *CallerProps

// Init is called during chaincode instantiation. The arguments passed
// to Init by the remote client includes:
//
//   0: Symbol of the token, e.g. "FUSD"
//   1: Name of the token, e.g. "Fabric USD: 1-1 peg to US Dollar"
//   2: Total token supply, e.g. "210000000"
//   3: Address of the initial owner of the tokens, e.g. "29cad..b6"
//
// Init could have alternatively used the invoker as the initial
// owner. The option of specifying a token owner allows the network to
// ensure the invoker does not have unncessary control over the entire
// token supply.
func (tcc *TokenChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// TODO: Validate args and handle upgrades
	args := stub.GetStringArgs()
	symbol := args[0]
	name := args[1]
	supply := stringToUint64(args[2])
	owner := args[3]

	t := Token{Symbol: symbol, Name: name, Decimals: 0, Supply: supply}
	b, err := json.Marshal(t)

	if err != nil {
		return shim.Error("Error marshalling token")
	}
	if err = stub.PutState("token", b); err != nil {
		shim.Error("Error writing token to ledger")
	}

	bal := Balance{Approved: nil, Available: supply}
	b, err = json.Marshal(bal)

	if err != nil {
		return shim.Error("Error marshalling balance")
	}
	if err = stub.PutState(owner, b); err != nil {
		shim.Error("Error writing owner's balance to ledger")
	}
	return shim.Success(nil)
}

// Invoke is called to update or query the state of the ledger. The
// arguments passed to Invoke by the remote client include:
//
//   0: The name of the function to Invoke. See 'SimpleToken'
//      interface for list of function names that can be supplied.
//   1..N: A list of arguments for the function defined in the
//      'SimpleToken' interface.
func (tcc *TokenChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	f, params := stub.GetFunctionAndParameters()
	var b []byte
	var err error

	// Retrieve token from ledger
	if b, err = stub.GetState("token"); err != nil {
		shim.Error("Error reading token from ledger")
	}
	tcc.token = &Token{}
	if err = json.Unmarshal(b, tcc.token); err != nil {
		shim.Error("Error unmarshaling token json")
	}

	// Initialize caller props for use in handlers
	cert, _ := cid.GetX509Certificate(stub)
	caller = &CallerProps{args: params, cert: cert, stub: stub}

	// Dispatch to appropriate handler based on supplied func name
	// TODO: Handle potential panics
	v := reflect.ValueOf(tcc).MethodByName(f + "Handler").Call([]reflect.Value{})
	return v[0].Interface().(pb.Response)
}

// TokenSupplyHandler fetches the total token supply of the
// underlying asset. The total supply is returned to the client in
// string form.
func (tcc *TokenChaincode) TokenSupplyHandler() pb.Response {
	supply, _ := tcc.token.TokenSupply()
	return shim.Success([]byte(strconv.FormatUint(supply, 10)))
}

// BalanceOfHandler fetches the balance available to the invoker for
// the underlying asset. The balance is returned to the client in
// string form.
func (tcc *TokenChaincode) BalanceOfHandler() pb.Response {
	// TODO: Validate args
	balance, err := tcc.token.BalanceOf(caller.args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(strconv.FormatUint(balance, 10)))
}

// TransferHandler transfers tokens from the invoker's address to the
// specified address. If the transfer is successful, the handler
// raises the 'Transferred' event and returns an empty payload.
func (tcc *TokenChaincode) TransferHandler() pb.Response {
	// TODO: Validate args
	to := caller.args[0]
	amount := stringToUint64(caller.args[1])
	if err := tcc.token.Transfer(to, amount); err != nil {
		return shim.Error(fmt.Sprintf("Failed to transfer tokens to %s: %s", to, err))
	}
	from := getInvokerAddress()
	_ = caller.stub.SetEvent("Transferred", newTransferredEvent(from, to, amount))
	return shim.Success(nil)
}

// ApproveHandler allows a spender to transfer tokens from the
// invoker's address to the specified address. If the approval was
// successful, the handler raises the 'Approved' event and returns an
// empty payload.
func (tcc *TokenChaincode) ApproveHandler() pb.Response {
	// TODO: Validate args
	spender := caller.args[0]
	amount := stringToUint64(caller.args[1])
	if err := tcc.token.Approve(spender, amount); err != nil {
		return shim.Error(fmt.Sprintf("Failed to approve token transfer to %s: %s", spender, err))
	}
	owner := getInvokerAddress()
	_ = caller.stub.SetEvent("Approved", newApprovedEvent(owner, spender, amount))
	return shim.Success(nil)
}

// TransferFromHandler transfers approved tokens from the owner's
// address to the specified address. The owner must have sufficient
// funds for the transfer. If the transfer was successful, the handler
// raises the 'Transferred' event and returns an empty payload.
func (tcc *TokenChaincode) TransferFromHandler() pb.Response {
	// TODO: Validate args
	from := caller.args[0]
	to := caller.args[1]
	amount := stringToUint64(caller.args[2])
	if err := tcc.token.TransferFrom(from, to, amount); err != nil {
		return shim.Error(fmt.Sprintf("Failed to transfer tokens from %s to %s: %s", from, to, err))
	}
	_ = caller.stub.SetEvent("Transferred", newTransferredEvent(from, to, amount))
	return shim.Success(nil)
}

// AllowanceHandler fetches the amount of tokens allowed for spending
// from a given owner's address by a given spender.
func (tcc *TokenChaincode) AllowanceHandler() pb.Response {
	// TODO: Validate args
	allowance, err := tcc.token.Allowance(caller.args[0], caller.args[1])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(strconv.FormatUint(allowance, 10)))
}

// newTransferredEvent returns a byte array representing a chaincode
// event for successful token transfers.
func newTransferredEvent(from string, to string, amount uint64) []byte {
	t := tokens.Transfer{From: from, To: to, Amount: amount}
	b, _ := json.Marshal(t)
	return b
}

// newApprovedEvent returns a byte array representing a chaincode
// event for successful approvals.
func newApprovedEvent(owner string, spender string, amount uint64) []byte {
	t := tokens.Approval{Owner: owner, Spender: spender, Amount: amount}
	b, _ := json.Marshal(t)
	return b
}

// getInvokerAddress gets a hex-based address representing the
// invoker's public key.
func getInvokerAddress() string {
	cert := security.NewX509Certificate(caller.cert)
	return cert.GetAddress()
}

// uint64ToBytes converts an unsigned integer to a byte array.
func uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

// uint64ToBytes converts a byte array to an unsigned integer.
func bytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

// uint64ToBytes converts a string to an unsigned integer.
func stringToUint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

func main() {
	tcc := new(TokenChaincode)
	if err := shim.Start(tcc); err != nil {
		fmt.Printf("Error starting TokenChaincode: %s", err)
	}
}
