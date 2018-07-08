{{define "NFToken"}}
pragma solidity ^0.4.23;

import \"MintableToken.sol\";

contract NFToken is MintableToken {

  constructor() public 
    MintableToken(\"{{.Name}}\", \"{{.Symbol}}\")
  { }

}
{{end}}