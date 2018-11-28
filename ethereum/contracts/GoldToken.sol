pragma solidity ^0.4.24;

import 'openzeppelin-solidity/contracts/token/ERC20/StandardToken.sol';

/**
 * @title Gold Token is a fictious asset backed token.
 * @dev Sample only. Not to be used in production.
 */
contract GoldToken is StandardToken {

  // A user friendly name of the token.
  string public name = "Gold Token. 1 Token: 1 Gram of Gold";

  // A unique token symbol (ticker) within a market
  string public symbol = "GLD";

  // The number of decimals for token fractions.
  uint8 public decimals = 18;

  // Total number of tokens in existence.
  uint public INITIAL_SUPPLY = 21000000;

  
  constructor() public {
    // See OpenZeppelin's ERC20Basic.sol
    // Total number of tokens in existence
    totalSupply_ = INITIAL_SUPPLY;

    // Assign all tokens to the contract creator.
    balances[msg.sender] = INITIAL_SUPPLY;
  }
}




