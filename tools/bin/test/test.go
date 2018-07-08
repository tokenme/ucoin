package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	//"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/params"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	//commonutils "github.com/tokenme/ucoin/utils"
	"github.com/ethereum/go-ethereum/common"
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
	if false {
		contractAddress, tx, _, err := utils.DeployNFToken(transactor, geth, "UCNF Test", "UCNF", "/opt/go/src/github.com/tokenme/ucoin/templates/contracts/erc721/*")
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("Contract: ", contractAddress.Hex(), "TX: ", txHash.Hex())

	}

	receipt, err := utils.TransactionReceipt(geth, ctx, "0x2e680cb890b4e65f1ccf3660afa423bdce862223b22d3a797164054f1caf4950")
	if err != nil {
		log.Fatalln(err)
	}
	if receipt == nil {
		return
	}
	spew.Dump(receipt)
	token, err := utils.NewNFToken("0x41bEC9F7E132758ccc3D2A0A5Bc6FAf16502C2B4", geth)
	if err != nil {
		log.Fatalln(err)
	}
	name, err := token.Name(nil)
	if err != nil {
		log.Fatalln(err)
	}
	owner, err := token.Owner(nil)
	if err != nil {
		log.Fatalln(err)
	}
	totalSupply, err := token.TotalSupply(nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Name: ", name, "Owner: ", owner.Hex(), "totalSupply: ", totalSupply.Uint64())
	if false {
		tx, err := token.Mint(transactor, common.HexToAddress("0xdd9dda410b7a9d03c2beedfbe89ef975deb41da8"), new(big.Int).SetUint64(1), "")
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
	}
}

func newGeth(ipcLocation string) (*ethclient.Client, error) {
	return ethclient.Dial(ipcLocation)
}
