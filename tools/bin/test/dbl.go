package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/params"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	//commonutils "github.com/tokenme/ucoin/utils"
	"log"
	"math/big"
)

func main() {
	geth, _ := newGeth("https://mainnet.infura.io/NlT37dDxuLT2tlZNw3It")
	ctx := context.Background()
	/*privKey, _, err := eth.GenerateAccount()
	if err != nil {
		log.Fatalln(err)
	}*/
	privKey := "073dabb8ca4b7fc058552f77761d43f3a385eb54d489ed4f0c665217152929cb"
	pubKey, err := eth.AddressFromHexPrivateKey(privKey)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Wallet: ", pubKey, ", Priv: ", privKey)
	transactor := eth.TransactorAccount(privKey)
	/*err = eth.TransactorUpdateWithEstimation(transactor, geth, ctx)
	  if err != nil {
	      log.Fatalln(err)
	  }
	transactor.GasLimit = 540000*/
	tokenAddress := "0x6748790CE50D9Df2a9e153E9F13Df05FB54cA3dB"
	if false {
		contractAddress, tx, _, err := utils.DeployToken(transactor, geth, "Life & luxury", "DBL", 18, 2000000000)
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("Contract: ", contractAddress.Hex(), "TX: ", txHash.Hex())
		return
	}

	token, err := utils.NewToken(tokenAddress, geth)
	if err != nil {
		log.Fatalln(err)
	}
	if true {
		name, symbol, decimals, initialSupply, totalSupply, totalHolders, circulatingSupply, balance, _, err := utils.TokenMeta(token, "")
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("name: ", name, ", symbol: ", symbol, ", decimals: ", decimals, ", initialSupply: ", initialSupply.String(), ", totalSupply: ", totalSupply.String(), ", totalHolders: ", totalHolders, ", circulatingSupply: ", circulatingSupply.Uint64(), "balance: ", balance.String())
	}

	if false {
		transactorOpts := eth.TransactorOptions{
			GasLimit: 540000,
		}
		eth.TransactorUpdate(transactor, transactorOpts, ctx)
		tx, err := token.Burn(transactor, big.NewInt(100000000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}

	if false {
		transactorOpts := eth.TransactorOptions{
			GasLimit: 540000,
		}
		eth.TransactorUpdate(transactor, transactorOpts, ctx)
		tx, err := token.Mint(transactor, common.HexToAddress(pubKey), big.NewInt(100000000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}

	if false {
		balance, err := utils.TokenBalanceOf(token, "0x251e3c2de440c185912ea701a421d80bbe5ee8c9")
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("balance: ", balance)
	}
	if false {
		transactorOpts := eth.TransactorOptions{
			GasLimit: 540000,
		}
		eth.TransactorUpdate(transactor, transactorOpts, ctx)
		//tx, err := token.TransferFrom(transactor, common.HexToAddress(pubKey), common.HexToAddress("0xf6b31e57df96dcd0e63da90623fa61434abf2202f3d44fca35e1d2c5d79ee5dd"), big.NewInt(10000))
		tx, err := token.Transfer(transactor, common.HexToAddress("0x251e3c2de440c185912ea701a421d80bbe5ee8c9"), big.NewInt(10000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}
	if false {
		tx, err := utils.TokenApprove(token, transactor, "", big.NewInt(10000000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}
	if true {
		receipt, err := utils.TransactionReceipt(geth, ctx, "0xc08c0dc731e9cd0d6ca30296aeaecde9d6653a3be87a0364e0b432da366649cc")
		if err != nil {
			log.Fatalln(err)
		}
		if receipt == nil {
			return
		}
		spew.Dump(receipt)
	}
}

func newGeth(ipcLocation string) (*ethclient.Client, error) {
	return ethclient.Dial(ipcLocation)
}
