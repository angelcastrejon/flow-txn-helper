package flowtxnhelper

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
)

func Account(flowClient access.Client, key string, address string) (flow.Address, *flow.AccountKey, crypto.Signer) {
	privateKey, err := crypto.DecodePrivateKeyHex(crypto.ECDSA_P256, key)
	HandleError(err)

	addr := flow.HexToAddress(address)
	acc, err := flowClient.GetAccount(context.Background(), addr)
	HandleError(err)

	accountKey := acc.Keys[0]
	signer, err := crypto.NewInMemorySigner(privateKey, accountKey.HashAlgo)
	HandleError(err)
	return addr, accountKey, signer
}

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetReferenceBlockId(flowClient access.Client) flow.Identifier {
	block, err := flowClient.GetLatestBlock(context.Background(), true)
	HandleError(err)

	return block.ID
}

func NewEmulatorFlowGRPCClient() (*grpc.Client, error) {
	c, err := grpc.NewClient(grpc.EmulatorHost)
	return c, err
}

func NewTestnetFlowGRPCClient() (*grpc.Client, error) {
	c, err := grpc.NewClient(grpc.TestnetHost)
	return c, err
}

func NewMainnetFlowGRPCClient() (*grpc.Client, error) {
	c, err := grpc.NewClient(grpc.MainnetHost)
	return c, err
}

func NewTestnetClient(key string, address string) (*grpc.Client, flow.Address, *flow.AccountKey, crypto.Signer, error) {
	c, err := NewTestnetFlowGRPCClient()
	addr, accountKey, signer := Account(c, key, address)
	return c, addr, accountKey, signer, err
}

func NewMainnetClient(key string, address string) (*grpc.Client, flow.Address, *flow.AccountKey, crypto.Signer, error) {
	c, err := NewMainnetFlowGRPCClient()
	addr, accountKey, signer := Account(c, key, address)
	return c, addr, accountKey, signer, err
}

func WaitForSeal(ctx context.Context, c access.Client, id flow.Identifier) *flow.TransactionResult {
	result, err := c.GetTransactionResult(ctx, id)
	HandleError(err)

	log.Printf("Waiting for transaction %s to be sealed...\n", id)

	for result.Status != flow.TransactionStatusSealed {
		time.Sleep(time.Second)
		fmt.Print(".")
		result, err = c.GetTransactionResult(ctx, id)
		HandleError(err)
	}

	fmt.Println()
	fmt.Printf("Transaction %s sealed\n", id)
	return result
}

func PrintTransaction(tx *flow.Transaction, err error) {
	HandleError(err)
	fmt.Println("Printing Transaction")
	fmt.Println("================================")
	fmt.Printf("ID().String(): %v\n", tx.ID().String())
	fmt.Printf("tx.Payer.String(): %v\n", tx.Payer.String())
	fmt.Printf("tx.ProposalKey.Address.String(): %v\n", tx.ProposalKey.Address.String())
	fmt.Printf("tx.Authorizers: %v\n", tx.Authorizers)
	fmt.Println("================================")
}

func PrintTransactionResult(txr *flow.TransactionResult, err error) {
	HandleError(err)
	fmt.Println("Printing Tx Result")
	fmt.Println("================================")
	fmt.Printf("\nStatus: %s", txr.Status.String())
	fmt.Printf("\nError: %v", txr.Error)
	fmt.Println("================================")
}

func PrintBlockTimeStamp(block *flow.Block, err error) string {
	HandleError(err)
	return block.Timestamp.String()
}

func PrintBlock(block *flow.Block, err error) {
	HandleError(err)
	fmt.Printf("\nID: %s\n", block.ID)
	fmt.Printf("height: %d\n", block.Height)
	fmt.Printf("timestamp: %s\n\n", block.Timestamp)
}

func CheckFileExists(fileName string) bool {
	_, error := os.Stat(fileName)
	return !errors.Is(error, os.ErrNotExist)
}

// RandomPrivateKey returns a randomly generated ECDSA P-256 private key.
func RandomPrivateKey() crypto.PrivateKey {
	seed := make([]byte, crypto.MinSeedLength)
	_, err := rand.Read(seed)
	HandleError(err)

	privateKey, err := crypto.GeneratePrivateKey(crypto.ECDSA_P256, seed)
	HandleError(err)

	return privateKey
}

// Generate a new AccountKey with a given weight
func NewAccountKey(pk crypto.PrivateKey, w int) *flow.AccountKey {
	return flow.NewAccountKey().
		FromPrivateKey(pk).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(w)
}
