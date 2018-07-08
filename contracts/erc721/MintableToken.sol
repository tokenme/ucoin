pragma solidity ^0.4.18;

import "ERC721Token.sol";
import "Ownable.sol";

/**
 * @title Mintable token
 * @dev Simple ERC721 Token example, with mintable token creation
 * @dev Issue: * https://github.com/OpenZeppelin/openzeppelin-solidity/issues/120
 * Based on code by TokenMarketNet: https://github.com/TokenMarketNet/ico/blob/master/contracts/MintableToken.sol
 */
contract MintableToken is ERC721Token, Ownable {
  event Mint(address indexed to, uint256 tokenId);
  event MintFinished();

  bool public mintingFinished = false;


  modifier canMint() {
    require(!mintingFinished);
    _;
  }

  modifier hasMintPermission() {
    require(msg.sender == owner);
    _;
  }

  constructor(string name, string symbol) public
    ERC721Token(name, symbol)
  { }

  /**
   * @dev Function to mint tokens
   * @param _to The address that will receive the minted tokens.
   * @param _tokenId uint256 ID of the token to be minted by the msg.sender.
   * @param _uri string URI to assign
   * @return A boolean that indicates if the operation was successful.
   */
  function mint(
    address _to,
    uint256 _tokenId,
    string _uri
  )
    hasMintPermission
    canMint
    public
    returns (bool)
  {
    super._mint(_to, _tokenId);
    super._setTokenURI(_tokenId, _uri);
    emit Mint(_to, _tokenId);
    return true;
  }

  /**
   * @dev Function to stop minting new tokens.
   * @return True if the operation was successful.
   */
  function finishMinting() onlyOwner canMint public returns (bool) {
    mintingFinished = true;
    emit MintFinished();
    return true;
  }

  /**
   * @dev Function to burn a specific token
   * Reverts if the token does not exist
   * @param _tokenId uint256 ID of the token being burned by the msg.sender
   */
  function burn(uint256 _tokenId) public {
    super._burn(ownerOf(_tokenId), _tokenId);
  }

  /**
   * @dev Function to set the token URI for a given token
   * Reverts if the token ID does not exist
   * @param _tokenId uint256 ID of the token to set its URI
   * @param _uri string URI to assign
   */

  function setTokenURI(uint256 _tokenId, string _uri) hasMintPermission public returns (bool) {
    super._setTokenURI(_tokenId, _uri);
    return true;
  }
}