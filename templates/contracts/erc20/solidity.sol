{{define "solidity"}}
{ "language": "Solidity",
  "sources":
  {
    "SafeMath": {
        "content": "{{template "safeMath"}}"
    },
    "Ownable":
    {
        "content": "{{template "ownable"}}"
    },
    "Pausable":
    {
        "content": "{{template "pausable"}}"
    },
    "ERC20Basic":
    {
        "content": "{{template "erc20Basic"}}"
    },
    "ERC20": {
        "content": "{{template "erc20"}}"
    },
    "BasicToken": {
        "content": "{{template "basicToken"}}"
    },
    "BurnableToken": {
        "content": "{{template "burnableToken"}}"
    },
    "StandardToken": {
        "content": "{{template "standardToken"}}"
    },
    "MintableToken": {
        "content": "{{template "mintableToken"}}"
    },
    "StandardBurnableToken": {
        "content": "{{template "standardBurnableToken"}}"
    },
    "PausableToken": {
        "content": "{{template "pausableToken"}}"
    },
    "Token": {
        "content": "{{template "token" .Token}}"
    }
  },
  "settings":
  {
    "optimizer": {
      "enabled": true,
      "runs": 200
    },
    "evmVersion": "homestead",
    "metadata": {
      "useLiteralContent": true
    },
    "outputSelection": {
      "*": {
        "*": [ "abi", "evm.bytecode.object"]
      }
    }
  }
}
{{end}}