pragma solidity ^0.4.24;

import 'openzeppelin-solidity/contracts/token/ERC20/StandardToken.sol';

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
  }

  // A map of agreement IDs and Agreements 
  mapping (bytes32 => Agreement) agreements;

  // Checks if a given agreement currently exists.
  modifier agreementExists(bytes32 agreementID) {
	 require(agreements[agreementID].counterparty != address(0))
  }
  
  /**
   * @dev No special construction logic.
   */
  constructor() public {
  }

  /** 
   * @dev lock creates a new swap agreement between the sender (owner) and
   * the counterparty.
   * @param counterpary The address of the counterparty in the swap.
   * @param image The SHA256 image of a known secret.
   * @param amount The amount of tokens to swap.
   * @param tokenContract The address of the underlying token contract to
   * invoke.
   * @param lockTime An agreed upon lock time during which the invoker is
   * unable to withdraw her tokens. 
   * @return The agreement id.
   */  
  function lock(
    address counterparty,
    bytes32 image,
    uint256 amount,
    address tokenContract,
    uint256 lockTime
  )
    external returns (string)
  {
    require(lockTime > 0);
    require (amount > 0);

    // Construct a unique ID for this agreement.
    bytes32 agreementID = sha256(abi.encodePacked(msg.sender, receiver, block.timestamp, image));

	 // TODO: Check if the contract already exists 
	 // TODO: image might require more space than bytes32
    agreements[agreementID] = Agreement(
       msg.sender,
       receiver,
       image,
       amount,
       tokenContract,
       block.timestamp + lockTime
    );
    
    StandardToken contract = StandardToken(tokenContract);

	 // Lock tokens by transferring from the initiator's account to
	 // this contract's address.
    assert(contract.transferFrom(msg.sender, address(this), amount));

	 emit Locked(agreementID, owner, counterparty, image, amount, expires);
  }

  /** 
   * @dev unlock releases tokens locked by the sender (owner). Tokens
   * can only be released once the lock time has elapsed.  
   * @param agreementID The ID of the agreement under which tokens 
   * were locked.
   * @return true if the unlock was successful and false otherwise.
   */  
  function unlock(
	 bytes32 agreementID
  )
	 agreementExists(agreementID)
	 external returns (bool)
  {
	 // Ensure tokens can only be unlocked after the lock time agreed
	 // between by both parties has expired.
	 require(block.timestamp > agreements[agreementID].expiry);
	 require(msg.sender == agreements[agreementID].sender);
	 require(sha256(abi.encodePacked(secret)) == agreements[agreementID].image);
	 
    StandardToken contract = StandardToken(agreements[agreementID].tokenContract);

	 // Unlock tokens by transferring from this contract's address to the
	 // initiator's (sender's) address.
	 assert(contract.transfer(msg.sender, agreements[agreementID].amount));

	 emit Unlocked(agreementID);
  }

  /** 
   * @dev claim allows the counterparty to claim tokens from the agreement
   * setup by the creator. 
   * @param agreementID The ID of the agreement under which tokens were 
   * locked.
   * @param secret The secret required to claim tokens.
   * @return true if the unlock was successful and false otherwise.
   */  
  function claim(
	 bytes32 agreementID,
	 bytes32 secret
  )
	 agreementExists(agreementID)
	 external returns (bool)
  {
	 // Ensure tokens can only be claimed before the lock time agreed
	 // between both parties has expired.
	 require(block.timestamp < agreements[agreementID].expiry);
	 require(msg.sender == agreements[agreementID].receiver);
	 require(sha256(abi.encodePacked(secret)) == agreements[agreementID].image);

    StandardToken contract = StandardToken(agreements[agreementID].tokenContract);

	 // Claim tokens by transferring from this contract's address to the
	 // initiator's (receiver's) address.
	 assert(contract.transfer(msg.sender, agreements[agreementID].amount));	 
  }
}
