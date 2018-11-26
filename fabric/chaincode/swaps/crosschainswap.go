package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// CrossChainSwap implements the HTLC interface.
//
// See lib/asset/htlc/HTLC
type CrossChainSwap struct {
}

// Agreement represents a swap contract between an owner of tokens and
// a counterparty. The construct of an agreement captures the
// underlying token contract, the amount of tokens to be swapped and
// the image of a secret required to claim tokens. An agreement
// expires after a pre-agreed period of time.
type Agreement struct {
	// The address of the token owner and creator of an agreement.
	Owner string `json:"owner"`

	// The address of the counterparty in the agreement who is allowed
	// to claim tokens before the expiry.
	Counterparty string `json:"counterparty"`

	// The image of a secret required to claim tokens.
	Image string `json:"image"`

	// The amount of tokens to be swapped in the agreement.
	Amount uint64 `json:"amount"`

	// The name of the token contract representing the tokens to be
	// swaped in the agreement.
	TokenContract string `json:"tokenContract"`

	// The time (wall clock) after which the agreement is considered to
	// have expired and tokens can be unlocked by the owner.
	Expiry int64 `json:"expiry"`
}

// Lock creates a new swap agreement between the token owner and a
// counterparty. The agreement includes the image of a known secret,
// the amount of tokens to swap, the name of the underlying token
// contract to invoke and an agreed upon lock time during which the
// invoker is unable to withdraw her tokens.
//
// The token owner must ensure an allowance to the amount specified in
// the agreement is made to the current contract's address. Invoking
// this function results in a transfer of funds from the owner's
// address to the current contract's address. The transfer is executed
// on the target contract by way of invoking the contract
// chaincode. The function returns the agreement ID.
func (ccs *CrossChainSwap) Lock(counterparty string, image string, amount uint64, tokenContract string, lockTime int64) (string, error) {
	var agreement *Agreement
	var err error
	agreementID := newAgreementID()
	// Verify if agreement ID is unique
	if agreement, err = ccs.getAgreement(agreementID); err != nil {
		return "", err
	}
	if agreement != nil {
		return "", fmt.Errorf("Agreement %s already exists", agreementID)
	}
	// Create new agreement and write to ledger
	invoker := getInvokerAddress()
	expiry := getExpiryTime(lockTime)
	agreement = &Agreement{
		Owner:         invoker,
		Counterparty:  counterparty,
		Image:         image,
		Amount:        amount,
		TokenContract: tokenContract,
		Expiry:        expiry}
	if err = ccs.putAgreement(agreementID, agreement); err != nil {
		return "", err
	}
	// TODO: Invoke token contract to check if the contract has
	// implemented support for 'chaincode addresses'.

	// Invoke token contract to 'lock' tokens to custom (chaincode) address.
	chaincodeAddress := getChaincodeAddress()
	args := argArray("TransferFrom", invoker, chaincodeAddress, strconv.FormatUint(amount, 10))
	result := caller.stub.InvokeChaincode(tokenContract, args, "")
	if result.Status != shim.OK {
		return "", fmt.Errorf("Error transferring tokens in contract %s: %s", tokenContract, result.Message)
	}
	return agreementID, nil
}

// Unlock releases tokens locked by the invoker (owner) under a given
// agreement id. Tokens can only be released once the lock time has
// elapsed.
//
// Invoking this function results in a transfer of funds from the
// current contract's address to the owner's address. The transfer is
// executed on the target contract by way of invoking the contract
// chaincode.
func (ccs *CrossChainSwap) Unlock(agreementID string) error {
	var agreement *Agreement
	var err error
	if agreement, err = ccs.getAgreement(agreementID); err != nil {
		return err
	}
	invoker := getInvokerAddress()
	if invoker != agreement.Owner {
		return fmt.Errorf("Attempting to unlock tokens belonging to %s", agreement.Owner)
	}
	if agreement.Expiry > time.Now().Unix() {
		return fmt.Errorf("Agreement is set to expire on %s", time.Unix(agreement.Expiry, 0).Format(time.RFC850))
	}
	// Invoke token contract to 'unlock' tokens from custom (chaincode) address.
	args := argArray("Transfer", agreement.Owner, strconv.FormatUint(agreement.Amount, 10))
	result := caller.stub.InvokeChaincode(agreement.TokenContract, args, "")
	if result.Status != shim.OK {
		return fmt.Errorf("Error transferring tokens in contract %s: %s", agreement.TokenContract, result.Message)
	}
	return nil
}

// Claim allows the counterparty to claim tokens from the agreement
// setup by the creator. The counterparty must provide the correct
// agreement id and secret to claim her tokens.
//
// Invoking this function results in a transfer of funds from the
// current contract's address to the counterparty's address. The
// transfer is executed on the target contract by way of invoking the
// contract chaincode.
func (ccs *CrossChainSwap) Claim(agreementID string, secret string) error {
	var agreement *Agreement
	var err error
	if agreement, err = ccs.getAgreement(agreementID); err != nil {
		return err
	}
	invoker := getInvokerAddress()
	if invoker != agreement.Counterparty {
		return fmt.Errorf("Attempting to claim tokens belonging to %s", agreement.Counterparty)
	}
	if agreement.Expiry < time.Now().Unix() {
		return fmt.Errorf("Agreement expired on %s", time.Unix(agreement.Expiry, 0).Format(time.RFC850))
	}
	if imageOf(secret) != agreement.Image {
		return fmt.Errorf("SHA256 of secret '%s' does not match image '%s'", secret, agreement.Image)
	}
	// Invoke token contract to 'unlock' tokens from custom (chaincode) address.
	args := argArray("Transfer", agreement.Counterparty, strconv.FormatUint(agreement.Amount, 10))
	result := caller.stub.InvokeChaincode(agreement.TokenContract, args, "")
	if result.Status != shim.OK {
		return fmt.Errorf("Error transferring tokens in contract %s: %s", agreement.TokenContract, result.Message)
	}
	return nil
}

// getAgreement returns the agreement with the specified ID from the ledger.
func (ccs *CrossChainSwap) getAgreement(agreementID string) (*Agreement, error) {
	var b []byte
	var err error
	if b, err = caller.stub.GetState(agreementID); err != nil {
		return nil, err
	}
	var agreement Agreement
	if b == nil {
		return nil, nil
	}
	if err = json.Unmarshal(b, &agreement); err != nil {
		return nil, err
	}
	return &agreement, nil
}

// putAgreement writes the given agreement to the ledger.
func (ccs *CrossChainSwap) putAgreement(agreementID string, agreement *Agreement) error {
	b, err := json.Marshal(&agreement)
	if err != nil {
		return err
	}
	if err = caller.stub.PutState(agreementID, b); err != nil {
		return err
	}
	return nil
}

// newAgreementID creates a unique agreement ID.
func newAgreementID() string {
	// The transaction ID is unique per transaction, per client.
	// This will serve as a good agreement ID.
	return caller.stub.GetTxID()
}

// imageOf returns the SHA256 hex representation of a given string.
func imageOf(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(h[:])
}

// argArray returns a slice over byte array, each element representing a
// byte representation of a string.
func argArray(s ...string) [][]byte {
	args := make([][]byte, len(s))
	for i, v := range s {
		args[i] = []byte(v)
	}
	return args
}
