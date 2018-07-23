pragma solidity ^0.4.23;

import "Ownable.sol";
import "ERC20.sol";


contract MultiSendERC20Dealer is Ownable {

  function multiSend(address _tokenAddr, address _dealer, uint256 _price, address[] recipients, uint256[] amounts) public onlyOwner payable {
    require(recipients.length == amounts.length);

    if (_price > 0) {
      _dealer.transfer(_price);
    }

    ERC20 token = ERC20(_tokenAddr);

    for (uint i = 0; i < recipients.length; i++) {
      address _to = recipients[i]; 
      uint256 _value = amounts[i];
      if (_to == address(0) || _to == msg.sender || _value == 0) {
        continue;
      }
      token.transferFrom(msg.sender, _to, _value);
    }
  }

  function destroy() public onlyOwner {
    selfdestruct(owner);
  }
}