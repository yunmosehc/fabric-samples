/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	
	"github.com/astaxie/beego"
)

var carStr string = ""

type HomeController struct {
        beego.Controller
}

func (this *HomeController) Get() {
        this.Data["carString"] = carStr
		this.TplName = "allCar.html"
}

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	contract := network.GetContract("fabcar")

	result, err := contract.EvaluateTransaction("queryAllCars")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	result, err = contract.SubmitTransaction("createCar", "CAR2", "联盟链开发实战",
		"https://ipfs.io/ipfs/QmQU2gS4gZ7TpiTECjDUxdQFd9bBBEWxDxPPfhLfYHVuei", "000002", "000000", "2020.10.20 18:20:30", "李白", "110100200101101201")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	result, err = contract.EvaluateTransaction("queryCar", "CAR2")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	_, err = contract.SubmitTransaction("changeCarOwner", "CAR2", "000003", "杜甫", "371510199002202838")
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}

	result, err = contract.EvaluateTransaction("queryCar", "CAR2")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	carStr = string(result)
	fmt.Println(string(result))
	
	beego.Router("/",&HomeController{})
	beego.Run()
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}
