{{define "solidity"}}
{ "language": "Solidity",
  "sources":
  {
    "SafeMath.sol": {
        "content": "{{template "SafeMath"}}"
    },
    "AddressUtils.sol": {
        "content": "{{template "AddressUtils"}}"
    },
    "Ownable.sol":
    {
        "content": "{{template "Ownable"}}"
    },
    "ERC165.sol":
    {
        "content": "{{template "ERC165"}}"
    },
    "SupportsInterfaceWithLookup.sol":
    {
        "content": "{{template "SupportsInterfaceWithLookup"}}"
    },
    "ERC721Receiver.sol": {
        "content": "{{template "ERC721Receiver"}}"
    },
    "ERC721Holder.sol": {
        "content": "{{template "ERC721Holder"}}"
    },
    "ERC721.sol": {
        "content": "{{template "ERC721"}}"
    },
    "ERC721Basic.sol": {
        "content": "{{template "ERC721Basic"}}"
    },
    "ERC721BasicToken.sol": {
        "content": "{{template "ERC721BasicToken"}}"
    },
    "ERC721Token.sol": {
        "content": "{{template "ERC721Token"}}"
    },
    "MintableToken.sol": {
        "content": "{{template "MintableToken"}}"
    },
    "NFToken": {
        "content": "{{template "NFToken" .Token}}"
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