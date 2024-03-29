pragma solidity ^0.4.24;

/** 
 * @title HTLC interface captures the protocol for a Hashed TimeLock
 * Contract (HTLC), sometimes called Hashed TimeLock Agreement (HTLA).
 * @dev An HTLC enables two parties, both of whom are members of two
 * seperate chains, to transfer ownership of tokens between them. The
 * exchange protocol begins by one party invoking Lock against a
 * certain number of tokens she owns using an image (of a secret) that
 * is then shared with the second party. The second party, likewise,
 * locks her tokens on the second chain. The first party can then
 * claim tokens on the second chain by revealing her secret, or unlock
 * her own tokens if the lock time has elapsed. The second party can
 * now use the disclosed secret to claim her share of the deal.
 */
contract HTLC {

  /** 
   * @dev lock creates a new swap agreement between the sender (owner) and
   * the counterparty.
   * @param counterparty The address of the counterparty in the swap.
   * @param image The SHA256 image of a known secret.
   * @param amount The amount of tokens to swap.
   * @param tokenContract The address of the underlying token contract to
   * invoke.
   * @param lockTime An agreed upon lock time during which the invoker is
   * unable to withdraw her tokens. 
   */
  function lock(
      address counterparty,
      bytes32 image,
      uint256 amount,
      address tokenContract,
      uint256 lockTime) public;

  /** 
   * @dev unlock releases tokens locked by the sender (owner). Tokens
   * can only be released once the lock time has elapsed.  
   * @param agreementID The ID of the agreement under which tokens 
   * were locked.
   */
  function unlock(bytes32 agreementID) public;

  /** 
   * @dev claim allows the counterparty to claim tokens from the agreement
   * setup by the creator. 
   * @param agreementID The ID of the agreement under which tokens were 
   * locked.
   * @param secret The secret required to claim tokens.
   */
  function claim(bytes32 agreementID, bytes secret) public;

  /** 
   * @dev Locked represents a lock event, raised when a new agreement is
   * created between the owner and a counterparty.
   */
  event Locked(
      bytes32 indexed agreementID,
      address indexed owner,
      address indexed counterparty,
      bytes32 image,
      uint256 amount,
      uint256 expiry);

  /**
   * @dev Unlocked represents an unlock event, raised when the owner releases
   * her tokens after the lock time has elapsed.  
   */
  event Unlocked(bytes32 indexed agreementID);

  /** 
   * @dev Claimed represents a claim event, raised when the counterparty
   * claims her tokens using the known secret.
   */
  event Claimed(bytes32 indexed agreementID, bytes secret);
}
