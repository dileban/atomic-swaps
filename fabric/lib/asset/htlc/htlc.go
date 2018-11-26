package htlc

// Locked represents a lock event, raised when a new agreement is
// created between the owner and a counterpary.
type Locked struct {
	AgreementID  string `json:"agreementId"`
	Owner        string `json:"owner"`
	CounterParty string `json:"counterparty"`
	Image        string `json:"image"`
	Amount       uint64 `json:"amount"`
	Expiry       int64  `json:"expiry"`
}

// Unlocked represents an unlock event, raised when the owner releases
// her tokens after the lock time has elapsed.
type Unlocked struct {
	AgreementID string `json:"agreementId"`
}

// Claimed represents a claim event, raised when the counterparty
// claims her tokens using the known secret.
type Claimed struct {
	AgreementID string `json:"agreementId"`
}

// HTLC interface captures the protocol for a Hashed TimeLock Contract
// (HTLC), sometimes called Hashed TimeLock Agreement (HTLA). An HTLC
// enables two parties, both of whom are members of two seperate
// chains, to transfer ownership of tokens between them. The exchange
// protocol begins by one party invoking Lock against a certain number
// of tokens she owns using an image (of a secret) that is then shared
// with the second party. The second party, likewise, locks her tokens
// on the second chain. The first party can then claim tokens on the
// second chain by revealing her secret, or unlock her own tokens if
// the lock time has elapsed. The second party can now use the
// disclosed secret to claim her share of the deal.
type HTLC interface {
	// Lock creates a new swap agreement between the invoker (owner)
	// and the counterparty. The agreement includes the image of a
	// known secret, the amount of tokens to swap, the name of the
	// underlying token contract to invoke and an agreed upon lock time
	// during which the invoker is unable to withdraw her tokens. Lock
	// returns the agreement id.
	Lock(counterparty string, image string, amount uint64, tokenContract string, lockTime int64) (string, error)

	// Unlock releases tokens locked by the invoker (owner) under a
	// given agreement id. Tokens can only be released once the
	// lock time has elapsed.
	Unlock(agreementID string) error

	// Claim allows the counterparty to claim tokens from the agreement
	// setup by the creator. The counterparty must provide the correct
	// agreement id and secret to claim her tokens.
	Claim(agreementID string, secret string) error
}
