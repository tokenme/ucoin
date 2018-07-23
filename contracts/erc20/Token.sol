pragma solidity ^0.4.23;

import "MintableToken.sol";
import "StandardBurnableToken.sol";
import "PausableToken.sol";

contract Token is MintableToken, StandardBurnableToken, PausableToken {

  string internal name_;
  string internal symbol_;
  uint8 internal decimals_ = 9;
  uint256 internal initialSupply_;

  constructor(string _name, string _symbol, uint8 _decimals, uint256 _initialSupply) public {
    name_ = _name;
    symbol_ = _symbol;
    decimals_ = _decimals;
    initialSupply_ = _initialSupply;
    totalSupply_ = initialSupply_;
    balances[msg.sender] = initialSupply_;
    emit Transfer(0x0, msg.sender, initialSupply_);
  }

  function name() external view returns (string) {
    return name_;
  }

  function symbol() external view returns (string) {
    return symbol_;
  }

  function decimals() external view returns (uint8) {
    return decimals_;
  }

  function initialSupply() external view returns (uint256) {
    return initialSupply_;
  }

  function circulatingSupply() public view returns (uint256) {
    require(totalSupply_ >= balances[owner]);
    return totalSupply_.sub(balances[owner]);
  }

  function meta(address owner) public view returns (string, string, uint8, uint256, uint256, uint256, uint256, uint256, uint256) {
    uint256 circulating = 0;
    if (totalSupply_ > balances[owner]) {
      circulating = totalSupply_.sub(balances[owner]);
    }
    uint256 balance = 0;
    if (owner != address(0)) {
      balance = balances[owner];
    } else if (msg.sender != address(0)) {
      balance = balances[msg.sender];
    }
    return (name_, symbol_, decimals_, initialSupply_, totalSupply_, totalTransfers_, totalHolders_, circulating, balance);
  }

  function batchTransfer(address[] addresses, uint256[] tokenAmount) public returns (bool) {
    require(addresses.length > 0 && addresses.length == tokenAmount.length);
    address _from = msg.sender;
    for (uint i = 0; i < addresses.length; i++) {
        address _to = addresses[i]; 
        uint256 _value = tokenAmount[i];
        require(_to != address(0));
        require(_value <= balances[_from]);
        require(_value <= allowed[_from][msg.sender]);
        
        super.transferFrom(_from, _to, _value);
    }
    return true;
  }

  function batchTransferFrom(address _from, address[] addresses, uint256[] tokenAmount) public returns (bool) {
    require(addresses.length > 0 && addresses.length == tokenAmount.length);
    for (uint i = 0; i < addresses.length; i++) {
        address _to = addresses[i]; 
        uint256 _value = tokenAmount[i];
        require(_to != address(0));
        require(_value <= balances[_from]);
        require(_value <= allowed[_from][msg.sender]);

        super.transferFrom(_from, _to, _value);
    }
    return true;
  }

}