package usecase

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	errInvalidType = errors.New("invalid input type")
)

type Nonce struct {
	ConfirmedNonce uint64
	PendingNonce   uint64
}

func newTransactionWithDynamicFee(
	chainId *big.Int,
	to *common.Address,
	nonce uint64,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	gas uint64,
	data []byte,
	value *big.Int,
	accessList types.AccessList,
) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:    chainId,
		To:         to,
		Nonce:      nonce,
		GasFeeCap:  gasFeeCap,
		GasTipCap:  gasTipCap,
		Gas:        gas,
		Value:      value,
		Data:       data,
		AccessList: accessList,
	})
}

func newLegacyTransaction(
	to *common.Address,
	nonce uint64,
	gasPrice *big.Int,
	gas uint64,
	data []byte,
	value *big.Int,
) *types.Transaction {
	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gas,
		To:       to,
		Value:    value,
		Data:     data,
	})
}

func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

func validNumber(input string) *big.Int {
	if input == "" {
		return big.NewInt(0)
	}
	matched, err := regexp.MatchString("([0-9])", input)
	if !matched || err != nil {
		return nil
	}
	amount, ok := math.ParseBig256(input)
	if !ok {
		return nil
	}
	return amount.Abs(amount)
}

// nolint
func contains(arr []*big.Int, value *big.Int) bool {
	for _, a := range arr {
		if a.Cmp(value) == 0 {
			return true
		}
	}
	return false
}

func parseAccessList(list string) (types.AccessList, error) {
	accessList := types.AccessList{}
	if list == "" {
		return accessList, nil
	}

	err := json.Unmarshal([]byte(list), &accessList)
	if err != nil {
		return types.AccessList{}, err
	}

	return accessList, nil
}
