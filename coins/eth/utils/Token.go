package utils

import (
	"bytes"
	"github.com/bububa/solc-go"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tidwall/gjson"
	"github.com/tokenme/ucoin/coins/eth"
	"math/big"
	"text/template"
)

func DeployToken(auth *bind.TransactOpts, backend bind.ContractBackend, name string, symbol string, decimals uint, totalSupply uint64, templatePath string) (common.Address, *types.Transaction, *eth.Token, error) {
	templ := template.Must(template.ParseGlob(templatePath))
	tpl := new(bytes.Buffer)
	token := map[string]interface{}{"Name": name, "Symbol": symbol, "Decimals": decimals, "TotalSupply": totalSupply}
	templ.ExecuteTemplate(tpl, "solidity", map[string]interface{}{"Token": token})
	compiledRet := solc.CompileStandard(tpl.String())
	//abi := gjson.Get(compiledRet, "contracts.Token.Token.abi").String()
	bytecode := gjson.Get(compiledRet, "contracts.Token.Token.evm.bytecode.object").String()
	return eth.DeployToken(auth, backend, bytecode)
}

func NewToken(addr string, backend bind.ContractBackend) (*eth.Token, error) {
	return eth.NewToken(common.HexToAddress(addr), backend)
}

func TokenAllowance(token *eth.Token, wallet string, contract string) (*big.Int, error) {
	return token.Allowance(nil, common.HexToAddress(wallet), common.HexToAddress(contract))
}

func TokenApprove(token *eth.Token, opts *bind.TransactOpts, spender string, value *big.Int) (*types.Transaction, error) {
	return token.Approve(opts, common.HexToAddress(spender), value)
}

func BalanceOfToken(geth *ethclient.Client, tokenAddress string, wallet string) (*big.Int, error) {
	token, err := eth.NewToken(common.HexToAddress(tokenAddress), geth)
	if err != nil {
		return nil, err
	}
	return TokenBalanceOf(token, wallet)
}

func TokenBalanceOf(token *eth.Token, wallet string) (*big.Int, error) {
	return token.BalanceOf(nil, common.HexToAddress(wallet))
}

func Transfer(token *eth.Token, opts *bind.TransactOpts, _to string, _value *big.Int) (*types.Transaction, error) {
	return token.Transfer(opts, common.HexToAddress(_to), _value)
}

func TokenDecimal(token *eth.Token, opts *bind.CallOpts) (int, error) {
	d, err := token.Decimals(opts)
	if err != nil {
		return 0, err
	}
	return int(d), nil
}
