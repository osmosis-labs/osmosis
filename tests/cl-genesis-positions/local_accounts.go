package main

import (
	"os/user"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/ignite/pkg/cosmosclient"

	"context"
	"fmt"
	"log"
)

const (
	expectedPoolId           uint64 = 1
	addressPrefix                   = "osmo"
	localosmosisFromHomePath        = "/.osmosisd-local"
	consensusFee                    = "1500uosmo"
	accountNamePrefix               = "lo-test"
	numPositions                    = 1_000
	randSeed                        = 1

	accountMax = 10
)

func GetLocalKeyringAccounts() []sdk.AccAddress {
	ctx := context.Background()

	clientHome := getClientHomePath()

	// Create a Cosmos igniteClient instance
	igniteClient, _ := cosmosclient.New(
		ctx,
		cosmosclient.WithAddressPrefix(addressPrefix),
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest),
		cosmosclient.WithHome(clientHome),
	)

	var err error
	igniteClient.AccountRegistry, err = cosmosaccount.New(
		cosmosaccount.WithKeyringBackend(cosmosaccount.KeyringTest),
		cosmosaccount.WithHome(clientHome),
	)
	if err != nil {
		panic(err)
	}

	accounts := make([]sdk.AccAddress, accountMax)
	for i := 1; i <= accountMax; i++ {
		accountName := fmt.Sprintf("%s%d", accountNamePrefix, i)
		account := getAccountAddressFromKeyring(igniteClient, accountName)
		_, err := igniteClient.AccountRegistry.Export(account.Name, "")
		if err != nil {
			panic(err)
		}
		accounts[i-1] = account.Info.GetAddress()
	}

	fmt.Println("retrieved accounts")

	return accounts
}

func getAccountAddressFromKeyring(igniteClient cosmosclient.Client, accountName string) cosmosaccount.Account {
	account, err := igniteClient.Account(accountName)
	if err != nil {
		log.Fatal(fmt.Errorf("did not find account with name (%s) in the keyring: %w", accountName, err))
	}
	return account
}

func getClientHomePath() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return ""
	}

	return currentUser.HomeDir + localosmosisFromHomePath
}
