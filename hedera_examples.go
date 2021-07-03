package main

import (
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/joho/godotenv"
)

func tranferTransaction(client *hedera.Client, fromAccount, toAccount hedera.AccountID, amount float64) (*hedera.TransactionReceipt, error) {
	transaction := hedera.NewTransferTransaction().
		AddHbarTransfer(fromAccount, hedera.HbarFrom(-amount, hedera.HbarUnits.Tinybar)).
		AddHbarTransfer(toAccount, hedera.HbarFrom(amount, hedera.HbarUnits.Tinybar))

	txResponse, err := transaction.Execute(client)
	if err != nil {
		return nil, err
	}

	transferReceipt, err := txResponse.GetReceipt(client)
	if err != nil {
		return nil, err
	}

	return &transferReceipt, nil
}

func createNewAccount(client *hedera.Client) (*hedera.AccountID, error) {
	newAccountPrivateKey, err := hedera.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	newAccountPublicKey := newAccountPrivateKey.PublicKey()

	//Create new account and assign the public key
	newAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(newAccountPublicKey).
		SetInitialBalance(hedera.HbarFrom(1000, hedera.HbarUnits.Tinybar)).
		Execute(client)

	receipt, err := newAccount.GetReceipt(client)
	if err != nil {
		return nil, err
	}

	return receipt.AccountID, nil
}

func getAccountBalance(client *hedera.Client, accountID hedera.AccountID) (hedera.AccountBalance, error) {
	query := hedera.NewAccountBalanceQuery().
		SetAccountID(accountID)
	return query.Execute(client)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("Unable to load environment variables from .env file. Error:\n%v\n", err))
	}

	//Grab your testnet account ID and private key from the .env file
	myAccountId, err := hedera.AccountIDFromString(os.Getenv("MY_ACCOUNT_ID"))
	if err != nil {
		panic(err)
	}
	myPrivateKey, err := hedera.PrivateKeyFromString(os.Getenv("MY_PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}

	fmt.Printf("The account ID is = %v\n", myAccountId)
	fmt.Printf("The private key is = %v\n", myPrivateKey)

	//Create your testnet client
	client := hedera.ClientForTestnet()
	client.SetOperator(myAccountId, myPrivateKey)

	newAccountId, err := createNewAccount(client)
	if err != nil {
		panic(err)
	}
	fmt.Printf("The new account ID is %v\n", *newAccountId)

	accountBalance, err := getAccountBalance(client, *newAccountId)
	if err != nil {
		panic(err)
	}
	fmt.Println("The account balance for the new account is", accountBalance.Hbars.AsTinybar())

	transferReceipt, err := tranferTransaction(client, myAccountId, *newAccountId, 10000)
	if err != nil {
		panic(err)
	}
	fmt.Printf("The transaction consensus status is %v\n", transferReceipt.Status)

	accountBalance, err = getAccountBalance(client, *newAccountId)
	if err != nil {
		panic(err)
	}
	fmt.Println("The account balance for the new account is", accountBalance.Hbars.AsTinybar())

	// get query cost
	balanceQuery := hedera.NewAccountBalanceQuery().
		SetAccountID(*newAccountId)
	cost, err := balanceQuery.GetCost(client)
	if err != nil {
		panic(err)
	}
	fmt.Println("The account balance query cost is:", cost.String())
}
