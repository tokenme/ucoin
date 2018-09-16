package main

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	commonutils "github.com/tokenme/ucoin/utils"
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
	//agentPrivKey := "161595354e033f54c86047ead669e45833a537dd972142e269ca5cf52d7fd9f3"
	agentPubKey := "0xec97a4010e2b3cf9c82033ac20fa985d31265fc2"
	transactor := eth.TransactorAccount(privKey)
	if false {
		transactor.Value = new(big.Int).Mul(big.NewInt(1000), big.NewInt(params.Ether))
		tx, err := eth.Transfer(transactor, geth, ctx, agentPubKey)
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
	}
	poolPrivKey, err := commonutils.AddressDecrypt("mQL1yxONjwBvms6n2wGNxATMIx0WdEpMklw528joC-NNzcYmJkLf72UWFPg2RGxDBYW4Da5NT8_4vZDFpe5-oWLPzhZiNyzw2g5mYnpPVok=", "ec548c11b663a2a548ec36eb6f45bca3847d2163", "RdH7aJSLWnx3i873EekoXksHa8G70OhEoROjYZPv3GZpLdNnPYcns9_YTptJT4heNegrnUpp3c8XtdrDIp4feFS8plymNgLiBMGZAPZxlpA=")
	if err != nil {
		log.Fatalln(err)
	}
	poolPubKey, err := eth.AddressFromHexPrivateKey(poolPrivKey)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Pool Wallet: ", poolPubKey)
	/*err = eth.TransactorUpdateWithEstimation(transactor, geth, ctx)
	  if err != nil {
	      log.Fatalln(err)
	  }
	  transactor.GasLimit = 540000*/
	tokenAddress := "0xa28ED02E990B1632Da50BCA8064c549133Fc6417"
	if false {
		contractAddress, tx, _, err := utils.DeployToken(transactor, geth, "TimeBankCoin", "TBC", 9, 1000000000)
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
		contractAddress, tx, _, err := utils.DeployNFToken(transactor, geth, "UCNF Test", "UCNF")
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
	if false {
		name, symbol, decimals, initialSupply, totalSupply, totalTransfers, totalHolders, circulatingSupply, balance, err := utils.TokenMeta(token, poolPubKey)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("name: ", name, ", symbol: ", symbol, ", decimals: ", decimals, ", initialSupply: ", initialSupply.Uint64(), ", totalSupply: ", totalSupply.Uint64(), ", totalTransfers: ", totalTransfers, ", totalHolders: ", totalHolders, ", circulatingSupply: ", circulatingSupply.Uint64(), "balance: ", balance.Uint64())
		totalTransfers, _ = token.TotalTransfers(nil)
		log.Println("TotalTransfers: ", totalTransfers)
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
		balance, err := utils.TokenBalanceOf(token, poolPubKey)
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
		amount := new(big.Int).Mul(big.NewInt(1000000), big.NewInt(params.Shannon))
		tx, err := token.Transfer(transactor, common.HexToAddress(poolPubKey), amount)
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
		tx, err := token.AddAgent(transactor, common.HexToAddress(agentPubKey))
		if err != nil {
			log.Fatalln(err)
		}
		txHash := tx.Hash()
		log.Println("TX: ", txHash.Hex())
		return
	}

	if false {
		transactorPool := eth.TransactorAccount(poolPrivKey)
		transactorOpts := eth.TransactorOptions{
			GasLimit: 540000,
		}
		eth.TransactorUpdate(transactorPool, transactorOpts, ctx)
		//tx, err := token.TransferFrom(transactor, common.HexToAddress(pubKey), common.HexToAddress("0xf6b31e57df96dcd0e63da90623fa61434abf2202f3d44fca35e1d2c5d79ee5dd"), big.NewInt(10000))
		amount := new(big.Int).Mul(big.NewInt(100), big.NewInt(params.Shannon))
		tx, err := token.TransferProxy(transactorPool, common.HexToAddress(poolPubKey), common.HexToAddress("0xea16d7bfcffcf2efd8b55d0f8e8263fe5420ed34"), amount)
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
		receipt, err := utils.TransactionReceipt(geth, ctx, "0x83ee0e92585de35ad72d81d4f88e8402682c3cb4150ecd01bea6cfec28f02fe0")
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
