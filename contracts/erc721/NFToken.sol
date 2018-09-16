pragma solidity ^0.4.23;

import "MintableToken.sol";

contract NFToken is MintableToken {
    
  constructor(string name, string symbol) public
    MintableToken(name, symbol)
  { }

}