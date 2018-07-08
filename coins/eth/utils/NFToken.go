package utils

import (
	"bytes"
	"github.com/bububa/solc-go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tidwall/gjson"
	"github.com/tokenme/ucoin/coins/eth"
	"math/big"
	"text/template"
)

func DeployNFToken(auth *bind.TransactOpts, backend bind.ContractBackend, name string, symbol string, templatePath string) (common.Address, *types.Transaction, *eth.NFToken, error) {
	templ := template.Must(template.ParseGlob(templatePath))
	tpl := new(bytes.Buffer)
	token := map[string]interface{}{"Name": name, "Symbol": symbol}
	templ.ExecuteTemplate(tpl, "solidity", map[string]interface{}{"Token": token})
	compiledRet := solc.CompileStandard(tpl.String())
	//abi := gjson.Get(compiledRet, "contracts.NFToken.NFToken.abi").String()
	bytecode := gjson.Get(compiledRet, "contracts.NFToken.NFToken.evm.bytecode.object").String()
	return eth.DeployNFToken(auth, backend, bytecode)
}

func NewNFToken(addr string, backend bind.ContractBackend) (*eth.NFToken, error) {
	return eth.NewNFToken(common.HexToAddress(addr), backend)
}

func NFTMint(token *eth.NFToken, opts *bind.TransactOpts, _to string, _tokenId *big.Int, _uri string) (*types.Transaction, error) {
	return token.Mint(opts, common.HexToAddress(_to), _tokenId, _uri)
}

func NFTTransferFrom(token *eth.NFToken, opts *bind.TransactOpts, _from string, _to string, _tokenId *big.Int) (*types.Transaction, error) {
	return token.TransferFrom(opts, common.HexToAddress(_from), common.HexToAddress(_to), _tokenId)
}
