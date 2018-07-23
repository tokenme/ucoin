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
	geth, _ := newGeth("https://geth.xibao100.com")
	ctx := context.Background()
	privKey := "37ef68ddea282962df570194c1126b7e494abd6810e29556740072092dbdadba"
	pubKey, err := eth.AddressFromHexPrivateKey(privKey)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Wallet: ", pubKey)
	transactor := eth.TransactorAccount(privKey)
	/*err = eth.TransactorUpdateWithEstimation(transactor, geth, ctx)
	  if err != nil {
	      log.Fatalln(err)
	  }
	  transactor.GasLimit = 540000*/
	tokenAddress := "0xD4de8684752B1D6bAEd3Ac97e4a808007d1b8134"
	if false {
		contractAddress, tx, _, err := utils.DeployToken(transactor, geth, "UCoin", "UC", 9, 1000000000)
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("Contract: ", contractAddress.Hex(), "TX: ", txHash.Hex())
		return
	}
	if false {
		contractAddress, tx, _, err := utils.DeployAirdropper(transactor, geth, tokenAddress)
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("Contract: ", contractAddress.Hex(), "TX: ", txHash.Hex())
		return
	}
	if false {
		contractAddress, tx, _, err := utils.DeployNFToken(transactor, geth, "UCNF Test", "UCNF", "/opt/go/src/github.com/tokenme/ucoin/templates/contracts/erc721/*")
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("Contract: ", contractAddress.Hex(), "TX: ", txHash.Hex())

	}

	dealerContract := "0xa432C3057122395F962b3848434Deab1447eac4E"
	token, err := utils.NewToken(tokenAddress, geth)
	if err != nil {
		log.Fatalln(err)
	}
	if true {
		name, symbol, decimals, initialSupply, totalSupply, totalHolders, circulatingSupply, balance, err := utils.TokenMeta(token, "")
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("name: ", name, ", symbol: ", symbol, ", decimals: ", decimals, ", initialSupply: ", initialSupply.Uint64(), ", totalSupply: ", totalSupply.Uint64(), ", totalHolders: ", totalHolders, ", circulatingSupply: ", circulatingSupply.Uint64(), "balance: ", balance.Uint64())
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
		balance, err := utils.TokenBalanceOf(token, pubKey)
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
		tx, err := token.TransferFrom(transactor, common.HexToAddress(pubKey), common.HexToAddress("0xf6b31e57df96dcd0e63da90623fa61434abf2202f3d44fca35e1d2c5d79ee5dd"), big.NewInt(10000))
		//tx, err := token.Transfer(transactor, common.HexToAddress("0xdd9dda410b7a9d03c2beedfbe89ef975deb41da8"), big.NewInt(10000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}

	balance, err := utils.TokenBalanceOf(token, "0xdd9dda410b7a9d03c2beedfbe89ef975deb41da8")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("balance: ", balance)
	if false {
		transactorOpts := eth.TransactorOptions{
			GasLimit: 540000,
		}
		eth.TransactorUpdate(transactor, transactorOpts, ctx)
		airdropper, err := utils.NewAirdropper(dealerContract, geth)
		if err != nil {
			log.Fatalln(err)
		}
		recipients := []common.Address{
			common.HexToAddress("0xdd9dda410b7a9d03c2beedfbe89ef975deb41da8"),
			common.HexToAddress("0x77ceb43f334b8f93251497ffcb8d6079fc8cf089"),
		}
		tokenAmounts := []*big.Int{
			big.NewInt(100),
			big.NewInt(200),
		}
		tx, err := utils.AirdropperDrop(airdropper, transactor, recipients, tokenAmounts)
		if err != nil {
			log.Fatalln(err)
		}
		spew.Dump(tx)
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}
	if false {
		tx, err := utils.TokenApprove(token, transactor, dealerContract, big.NewInt(10000000))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}
	if true {
		receipt, err := utils.TransactionReceipt(geth, ctx, "0x0d4e3bb4b7adb084735ac1407796c1a50ceb056210ff680cbe262a21aa5173b1")
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
