pragma solidity ^0.4.24;

import 'openzeppelin-solidity/contracts/token/ERC20/StandardToken.sol';
import './HTLC.sol';

/**
 * @title Contract for Atomic Swaps.
 * @dev AtomicSwap implements the HTLC interface.
 */
contract AtomicSwap is HTLC {

  // NOTE: Currently not used. Intended to support
  //       swaps on non-fungibles.
  enum TokenType {
    ERC20,
    ERC271
  }

  // Status of an agreement.
  enum Status {
    Locked,
    Unlocked,
    Claimed
  }

  // Agreement represents a swap contract between an owner of tokens
  // and a counterparty. The construct of an agreement captures the
  // underlying token contract, the amount of tokens to be swapped and
  // the image of a secret required to claim tokens. An agreement
  // expires after a pre-agreed period of time.
  struct Agreement {
    // The address of the token owner and creator of an agreement.
    address owner;
    // The address of the counterparty in the agreemetn who is allowed
    // to claim tokens before the expiry.
    address counterparty;
    // The image of a secret requred to claim tokens.
    bytes32 image;
    // The amount of tokens to be swapped in the agreement.
    uint256 amount;
    // The address of the token contract representing the tokens to be
    // swaped in the agreement.
    address tokenContract;
    // The time (wall clock) after which the agreement is considered to
    // have expired and tokens can be unlocked by the owner.
    uint256 expiry;
    // The status of an agreement.
    Status status;
  }

  // A map of agreement IDs and Agreements
  mapping (bytes32 => Agreement) agreements;

  // Checks if a given agreement currently exists.
  modifier agreementExists(bytes32 agreementID) {
    require(agreements[agreementID].counterparty != address(0),
            "Agreement does not exist");
    _;
  }

  // Checks if a given agreement currently 'Locked'
  modifier agreementLocked(bytes32 agreementID) {
    require(agreements[agreementID].status == Status.Locked,
            "Agreement must have status Locked");
    _;
  }

  /**
   * @dev No special construction logic.
   */
  constructor() public {
  }

  /**
   * @dev lock creates a new swap agreement between the sender (owner)
   * and the counterparty. The status of the agreement is initialized
   * to Status.Locked and the 'Locked' event is fired.
   * @param counterparty The address of the counterparty in the swap.
   * @param image The SHA256 image of a known secret.
   * @param amount The amount of tokens to swap.
   * @param tokenContract The address of the underlying token contract
   * to invoke.
   * @param lockTime An agreed upon lock time during which the invoker
   * is unable to withdraw her tokens.
   */
  function lock(
    address counterparty,
    bytes32 image,
    uint256 amount,
    address tokenContract,
    uint256 lockTime
  )
    public
  {
    require(counterparty != address(0),
            "Counterparty address is not valid");
    require(tokenContract != address(0),
            "Token contract address is not valid");
    require(lockTime > 0,
            "Lock time must be greater than 0");
    require(amount > 0,
            "Amount must be greater than 0");

    // Construct a unique agreement ID and calculate expiry.
    bytes32 agreementID = sha256(abi.encodePacked(msg.sender, counterparty, block.timestamp, image));
    uint256 expiry = block.timestamp + lockTime;

    // Ensure agreement ID doesn't exist already, check for zero-ed struct value
    require(agreements[agreementID].counterparty == address(0));

    // Create and assign new agreement to map
    agreements[agreementID] = Agreement(
       msg.sender,
       counterparty,
       image,
       amount,
       tokenContract,
       expiry,
       Status.Locked
    );

    StandardToken st = StandardToken(tokenContract);

    // Lock tokens by transferring from the initiator's account to
    // this contract's address.
    assert(st.transferFrom(msg.sender, address(this), amount));

    // Notify listeners.
    emit Locked(agreementID, msg.sender, counterparty, image, amount, expiry);
  }

  /**
   * @dev unlock releases tokens locked by the sender (owner). Tokens
   * can only be released once the lock time has elapsed. The
   * agreement status is updated to Status.Unlocked and the 'Unlocked'
   * event is fired.
   * @param agreementID The ID of the agreement under which tokens
   * were locked.
   */
  function unlock(
    bytes32 agreementID
  )
    agreementExists(agreementID)
    agreementLocked(agreementID)
    public
  {
    // Ensure tokens can only be unlocked after the lock time agreed
    // between by both parties has expired.
    require(block.timestamp > agreements[agreementID].expiry,
            "Agreement has not expired");
    require(msg.sender == agreements[agreementID].owner,
            "Agreement can only be unlocked by owner");

    StandardToken st = StandardToken(agreements[agreementID].tokenContract);

    // Unlock tokens by transferring from this contract's address to the
    // initiator's (sender's) address.
    assert(st.transfer(msg.sender, agreements[agreementID].amount));

    // Update agreement status and notify listeners.
    agreements[agreementID].status = Status.Unlocked;
    emit Unlocked(agreementID);
  }

  /**
   * @dev claim allows a counterparty to claim tokens from the
   * agreement setup by the creator. The agreement status is updated
   * to Status.Claimed and the 'Claimed' event is fired.
   * @param agreementID The ID of the agreement under which
   * tokens were locked.
   * @param secret The secret required to claim tokens.
   */
  function claim(
    bytes32 agreementID,
    bytes secret
  )
    agreementExists(agreementID)
    agreementLocked(agreementID)
    public
  {
    // Ensure tokens can only be claimed before the lock time agreed
    // between both parties has expired.
    require(block.timestamp < agreements[agreementID].expiry,
            "Agreement has expired");
    require(msg.sender == agreements[agreementID].counterparty,
            "Agreement can only be claimed by owner");
    require(sha256(abi.encodePacked(secret)) == agreements[agreementID].image,
            "Secret does not match");

    StandardToken st = StandardToken(agreements[agreementID].tokenContract);

    // Claim tokens by transferring from this contract's address to the
    // initiator's (counterparty's) address.
    assert(st.transfer(msg.sender, agreements[agreementID].amount));

    // Update agreement status and notify listeners.
    agreements[agreementID].status = Status.Claimed;
    emit Claimed(agreementID, secret);
  }

  /**
   * @dev A convenience method for returning the status of a given agreement.
   * @param agreementID The ID of the agreement under which tokens were
   * locked.
   * @return The status of the agreement as an ordinal value of enum Status.
   */
  function getStatus(
    bytes32 agreementID
  )
    agreementExists(agreementID)
    public view returns (uint)
  {
    return uint(agreements[agreementID].status);
  }
}
