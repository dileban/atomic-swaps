package main

import (
	"encoding/json"
	"fmt"
)

// Token implements SimpleToken interface and represents basic
// properties of the token, such as symbol, name and total supply.
//
// See lib/asset/fungible/SimpleToken
type Token struct {
	// Symbol is a short ticker symbol for the token, e.g. "FUSD".
	Symbol string `json:"symbol"`

	// Name of the token being represented, e.g. "Fabric USD".
	Name string `json:"name"`

	// Decimals allows for fractional tranfers.
	// [NOTE]: Currently not implemented.
	Decimals uint64 `json:"decimals"`

	// Supply is the total token supply, fixed at the time of creation.
	Supply uint64 `json:"supply"`
}

// Balance represents the tokens available for spending by an 'owner'
// as well as a list of approved transfers by other 'spenders' from
// the 'owners' account.
type Balance struct {
	// Approved is the amount approved for transferring by a 'spender'
	// if sufficient balance is available.
	Approved map[string]uint64 `json:"approved"`

	// Available is the current token balance avaiable for spending by
	// the 'owner'.
	Available uint64 `json:"available"`
}

// TokenSupply returns the total token supply.
func (t *Token) TokenSupply() (uint64, error) {
	return t.Supply, nil
}

// BalanceOf returns the token balance of the specified owner.
func (t *Token) BalanceOf(owner string) (uint64, error) {
	var bal *Balance
	var err error
	if bal, err = t.getBalance(owner); err != nil {
		return 0, err
	}
	return bal.Available, nil
}

// Transfer transfers tokens from the invoker to the specified
// address. The invoker must have sufficient funds to transfer. The
// function returns and error if the transfer unsuccessful.
func (t *Token) Transfer(to string, amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("Attempting to transfer zero amount")
	}
	// Get invoker's current balance
	sender := getInvokerAddress()
	bal, err := t.getBalance(sender)
	if err != nil {
		return err
	}
	// Check for sufficient funds
	if bal.Available < amount {
		return fmt.Errorf("Insufficient balance for %s", sender)
	}
	// Update sender's balance
	bal.Available -= amount
	if err = t.putBalance(sender, bal); err != nil {
		return err
	}
	// Get receiver's current balance
	if bal, err = t.getBalance(to); err != nil {
		return err
	}
	// Update  receiver's balance
	bal.Available += amount
	if err = t.putBalance(to, bal); err != nil {
		return err
	}
	return nil
}

// Approve will allow 'spender' to transfer 'amount' tokens from the
// invoker (owner) by calling TransferFrom. Calling Approve multiple
// times will overwrite the previous amount.
func (t *Token) Approve(spender string, amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("Attempting to approve zero amount")
	}
	// Get invoker's current balance
	sender := getInvokerAddress()
	bal, err := t.getBalance(sender)
	if err != nil {
		return err
	}
	if bal.Approved == nil {
		bal.Approved = make(map[string]uint64)
	}
	// Overwrite previously approved amount if any
	bal.Approved[spender] = amount
	return nil
}

// TransferFrom allows the invoker to transfer up to 'amount' tokens
// from the owner's ('from') account to the receiver's ('to')
// account. The invoker is allowed to call TransferFrom multiple times
// as long as there are sufficient funds.
func (t *Token) TransferFrom(from string, to string, amount uint64) error {
	if amount == 0 {
		return fmt.Errorf("Attempting to transfer zero amount")
	}
	// Get 'from's current balance
	bal, err := t.getBalance(from)
	if err != nil {
		return err
	}
	// Check if sender is eligble to transfer tokens
	sender := getInvokerAddress()
	if bal.Approved[sender] < amount {
		return fmt.Errorf("Insufficent balance approved for %s", sender)
	}
	if bal.Available < amount {
		return fmt.Errorf("Insufficient balance for %s", sender)
	}
	// Update 'from's balance
	bal.Approved[sender] -= amount
	bal.Available -= amount
	if err = t.putBalance(from, bal); err != nil {
		return err
	}
	// Update 'to's balance
	if bal, err = t.getBalance(to); err != nil {
		return err
	}
	bal.Available += amount
	if err = t.putBalance(to, bal); err != nil {
		return err
	}
	return nil
}

// Allowance returns the amount of tokens approved by an owner for
// spending by a given 'spender'.
func (t *Token) Allowance(owner string, spender string) (uint64, error) {
	// Get 'owner's current balance
	bal, err := t.getBalance(owner)
	if err != nil {
		return 0, err
	}
	return bal.Approved[spender], nil
}

// getBalance returns owner's current balance from the ledger.
func (t *Token) getBalance(owner string) (*Balance, error) {
	var b []byte
	var err error
	if b, err = caller.stub.GetState(owner); err != nil {
		return nil, err
	}
	var bal Balance
	if b == nil {
		return &bal, nil
	}
	if err = json.Unmarshal(b, &bal); err != nil {
		return nil, err
	}
	return &bal, nil
}

// putBalance writes owner's balance to the ledger.
func (t *Token) putBalance(owner string, bal *Balance) error {
	b, err := json.Marshal(bal)
	if err != nil {
		return err
	}
	if err = caller.stub.PutState(owner, b); err != nil {
		return err
	}
	return nil
}
