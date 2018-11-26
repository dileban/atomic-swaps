package asset

// TODO: Look into language neutral options

// Transfer represents a transfer event, raised when the transfer of
// tokens from an owner to a recipient is successful.
type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount uint64 `json:"amount"`
}

// Approval represents an approval event, raised when an amount of
// tokens has been approved for spending by a 'spender'.
type Approval struct {
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
	Amount  uint64 `json:"amount"`
}

// SimpleToken interface is modeled after Ethereum's ERC20 standard.
//
// See: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20.md
type SimpleToken interface {
	// TokenSupply returns the total token supply.
	TokenSupply() (uint64, error)

	// BalanceOf returns the token balance of the specified owner.
	BalanceOf(owner string) (uint64, error)

	// Transfer transfers tokens from the invoker to the specified
	// address. The invoker must have sufficient funds to transfer.
	Transfer(to string, amount uint64) error

	// Approve will allow 'spender' to transfer 'amount' tokens from
	// the invoker (owner) by calling TransferFrom. Calling Approve
	// multiple times overwrites the previous approved amount.
	Approve(spender string, amount uint64) error

	// TransferFrom allows the invoker to transfer up to 'amount'
	// tokens from the owner's ('from') account to the receiver's
	// ('to') account. The invoker is allowed to call TransferFrom
	// multiple times as long as there are sufficient funds.
	TransferFrom(from string, to string, amount uint64) error

	// Allowance returns the amount of tokens approved by an owner for
	// spending by a given 'spender'.
	Allowance(owner string, spender string) (uint64, error)
}
