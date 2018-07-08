{{define "token"}}
pragma solidity ^0.4.23;

import \"MintableToken\";
import \"StandardBurnableToken\";
import \"PausableToken\";

contract Token is MintableToken, StandardBurnableToken, PausableToken {

  string public constant name = \"{{.Name}}\";
  string public constant symbol = \"{{.Symbol}}\";
  uint8 public constant decimals = {{.Decimals}};
  uint256 public constant INITIAL_SUPPLY = {{.TotalSupply}} * (10 ** uint256(decimals));

  constructor() public {
    totalSupply_ = INITIAL_SUPPLY;
    balances[msg.sender] = INITIAL_SUPPLY;
    emit Transfer(0x0, msg.sender, INITIAL_SUPPLY);
  }

  function circulatingSupply() public view returns (uint256) {
    require(totalSupply_ > balances[owner]);
    return totalSupply_.sub(balances[owner]);
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
{{end}}