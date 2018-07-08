{{define "basicToken"}}
pragma solidity ^0.4.23;

import \"ERC20Basic\";
import \"SafeMath\";

/**
 * @title Basic token
 * @dev Basic version of StandardToken, with no allowances.
 */
contract BasicToken is ERC20Basic {
  using SafeMath for uint256;

  mapping(address => uint256) balances;

  uint256 totalSupply_;

  uint256 totalHolders_;

  uint256 totalTransfers_;

  /**
  * @dev total number of tokens in existence
  */
  function totalSupply() public view returns (uint256) {
    return totalSupply_;
  }

  /**
  * @dev total number of token holders in existence
  */
  function totalHolders() public view returns (uint256) {
    return totalHolders_;
  }

  function _increaseHolders(address _addr) internal {
    if (balances[_addr] > 0) {
      return;
    }
    totalHolders_ = totalHolders_.add(1);
  }

  function _decreaseHolders(address _addr) internal {
    if (balances[_addr] > 0 && totalHolders_ > 0) {
      return;
    }
    totalHolders_ = totalHolders_.sub(1);
  }

  /**
  * @dev total number of token transfers
  */
  function totalTransfers() public view returns (uint256) {
    return totalTransfers_;
  }

  /**
  * @dev transfer token for a specified address
  * @param _to The address to transfer to.
  * @param _value The amount to be transferred.
  */
  function transfer(address _to, uint256 _value) public returns (bool) {
    require(_to != address(0));
    require(_value <= balances[msg.sender]);

    balances[msg.sender] = balances[msg.sender].sub(_value);

    _decreaseHolders(msg.sender);
    _increaseHolders(_to);

    balances[_to] = balances[_to].add(_value);

    totalTransfers_ = totalTransfers_.add(1);

    emit Transfer(msg.sender, _to, _value);
    return true;
  }

  /**
  * @dev Gets the balance of the specified address.
  * @param _owner The address to query the the balance of.
  * @return An uint256 representing the amount owned by the passed address.
  */
  function balanceOf(address _owner) public view returns (uint256) {
    return balances[_owner];
  }

}
{{end}}
