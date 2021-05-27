// Copyright © 2018,2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ethbinding

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// EthAPI is a subset of the go-ethereum LGPL API exposed via a go plugin
type EthAPI interface {
	HexDecode(hex string) ([]byte, error)
	HexToAddress(hex string) Address
	BytesToAddress(b []byte) Address
	IsHexAddress(s string) bool
	HexToHash(hex string) Hash
	EncodeBig(bigint *big.Int) string
	FromHex(hex string) []byte
	ABITypeFor(typeName string) (ABIType, error)
	ABITypeKnown(typeName string) ABIType
	NewType(typeName string, internalType string) (typ abi.Type, err error)
	ABIEventSignature(event *ABIEvent) string
	ABIMarshalingToABIRuntime(marshalable ABIMarshaling) (*RuntimeABI, error)
	ABIArgumentsMarshalingToABIArguments(marshalable []ABIArgumentMarshaling) (ABIArguments, error)
	ABIElementMarshalingToABIEvent(marshalable *ABIElementMarshaling) (event *ABIEvent, err error)
	ABIElementMarshalingToABIMethod(m *ABIElementMarshaling) (method *ABIMethod, err error)
	JSON(reader io.Reader) (abi.ABI, error)
	SolidityVersion(solc string) (*Solidity, error)
	ParseCombinedJSON(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error)
	Dial(rawurl string) (*RPCClient, error)
	NewMethod(name string, rawName string, funType ABIFunctionType, mutability string, isConst, isPayable bool, inputs ABIArguments, outputs ABIArguments) ABIMethod
	NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction
	NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction
	NewEIP155Signer(chainID *big.Int) EIP155Signer
	ParseBig256(s string) (*big.Int, bool)
	S256(x *big.Int) *big.Int
	GenerateKey() (*ecdsa.PrivateKey, error)
	PubkeyToAddress(p ecdsa.PublicKey) Address
	FromECDSA(priv *ecdsa.PrivateKey) []byte
	HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error)
	NewStream(r io.Reader, inputLimit uint64) *rlp.Stream
	SignTx(tx *types.Transaction, s types.Signer, prv *ecdsa.PrivateKey) (*types.Transaction, error)
}

// EthAPIShim is an implementation of the shim
var EthAPIShim EthAPI = &ethAPIShim{}

type ethAPIShim struct{}

// This module provides access to some utils from the common package, with
// type mapping

// HexDecode decodes an 0x prefixed hex string
func (e *ethAPIShim) HexDecode(hex string) ([]byte, error) {
	return hexutil.Decode(hex)
}

// HexToAddress convert hex to an address
func (e *ethAPIShim) HexToAddress(hex string) Address {
	var addr Address = common.HexToAddress(hex)
	return addr
}

// BytesToAddress converts bytes to address
func (e *ethAPIShim) BytesToAddress(b []byte) Address {
	return common.BytesToAddress(b)
}

func (e *ethAPIShim) IsHexAddress(s string) bool {
	return common.IsHexAddress(s)
}

// HexToHash convert hex to an address
func (e *ethAPIShim) HexToHash(hex string) Hash {
	var hash Hash = common.HexToHash(hex)
	return hash
}

func (e *ethAPIShim) EncodeBig(bigint *big.Int) string {
	return hexutil.EncodeBig(bigint)
}

// FromHex returns the bytes represented by the hexadecimal string s
func (e *ethAPIShim) FromHex(hex string) []byte {
	return common.FromHex(hex)
}

// ABITypeFor gives you a type for a string
func (e *ethAPIShim) ABITypeFor(typeName string) (ABIType, error) {
	var t ABIType
	t, err := abi.NewType(typeName, "", []abi.ArgumentMarshaling{})
	return t, err
}

// ABITypeKnown gives you a type for a string you are sure is known
func (e *ethAPIShim) ABITypeKnown(typeName string) ABIType {
	var t ABIType
	t, _ = abi.NewType(typeName, "", []abi.ArgumentMarshaling{})
	return t
}

// NewType creates a new reflection type of abi type given in t
func (e *ethAPIShim) NewType(typeName string, internalType string) (typ abi.Type, err error) {
	return abi.NewType(typeName, internalType, []abi.ArgumentMarshaling{})
}

// ABIEventSignature returns a signature for an ABI event
func (e *ethAPIShim) ABIEventSignature(event *ABIEvent) string {
	typeStrings := make([]string, len(event.Inputs))
	for i, input := range event.Inputs {
		typeStrings[i] = input.Type.String()
	}
	return fmt.Sprintf("%v(%v)", event.RawName, strings.Join(typeStrings, ","))
}

// ABIMarshalingToABIRuntime takes a serialized form ABI and converts it into RuntimeABI
// This is not performance optimized, so the RuntimeABI once generated should be used
// for runtime processing
func (e *ethAPIShim) ABIMarshalingToABIRuntime(marshalable ABIMarshaling) (*RuntimeABI, error) {
	var runtime RuntimeABI
	b, _ := json.Marshal(&marshalable)
	err := json.Unmarshal(b, &runtime)
	return &runtime, err
}

// ABIArgumentsMarshalingToABIArguments converts ABI serialized reprsentations of arguments
// to the processed type information
func (e *ethAPIShim) ABIArgumentsMarshalingToABIArguments(marshalable []ABIArgumentMarshaling) (ABIArguments, error) {
	arguments := make(ABIArguments, len(marshalable))
	var err error
	for i, mArg := range marshalable {
		var components []abi.ArgumentMarshaling
		if mArg.Components != nil {
			b, _ := json.Marshal(&mArg.Components)
			_ = json.Unmarshal(b, &components)
		}
		arguments[i].Type, err = abi.NewType(mArg.Type, mArg.InternalType, components)
		if err != nil {
			return nil, err
		}
		arguments[i].Name = mArg.Name
		arguments[i].Indexed = mArg.Indexed
	}
	return arguments, nil
}

// ABIElementMarshalingToABIEvent converts a de-serialized event with full type information,
// per the original ABI, into a runtime ABIEvent with a processed type
func (e *ethAPIShim) ABIElementMarshalingToABIEvent(marshalable *ABIElementMarshaling) (event *ABIEvent, err error) {
	inputs, err := e.ABIArgumentsMarshalingToABIArguments(marshalable.Inputs)
	if err == nil {
		nEvent := abi.NewEvent(marshalable.Name, marshalable.Name, marshalable.Anonymous, inputs)
		event = &nEvent
	}
	return
}

// ABIElementMarshalingToABIMethod converts a de-serialized method with full type information,
// per the original ABI, into a runtime ABIEvent with a processed type
func (e *ethAPIShim) ABIElementMarshalingToABIMethod(m *ABIElementMarshaling) (method *ABIMethod, err error) {
	var inputs, outputs ABIArguments
	inputs, err = e.ABIArgumentsMarshalingToABIArguments(m.Inputs)
	if err == nil {
		outputs, err = e.ABIArgumentsMarshalingToABIArguments(m.Outputs)
		if err == nil {
			nMethod := abi.NewMethod(m.Name, m.Name, abi.Function, m.StateMutability, m.Constant, m.Payable, inputs, outputs)
			method = &nMethod
		}
	}
	return
}

// JSON returns a parsed ABI interface and error if it failed
func (e *ethAPIShim) JSON(reader io.Reader) (abi.ABI, error) {
	return abi.JSON(reader)
}

func (e *ethAPIShim) SolidityVersion(solc string) (*Solidity, error) {
	return compiler.SolidityVersion(solc)
}

func (e *ethAPIShim) ParseCombinedJSON(combinedJSON []byte, source string, languageVersion string, compilerVersion string, compilerOptions string) (map[string]*Contract, error) {
	return compiler.ParseCombinedJSON(combinedJSON, source, languageVersion, compilerVersion, compilerOptions)
}

func (e *ethAPIShim) Dial(rawurl string) (*RPCClient, error) {
	return rpc.Dial(rawurl)
}

func (e *ethAPIShim) NewMethod(name string, rawName string, funType ABIFunctionType, mutability string, isConst, isPayable bool, inputs ABIArguments, outputs ABIArguments) ABIMethod {
	return abi.NewMethod(name, rawName, funType, mutability, isConst, isPayable, inputs, outputs)
}

func (e *ethAPIShim) NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
}

func (e *ethAPIShim) NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data)
}

func (e *ethAPIShim) NewEIP155Signer(chainID *big.Int) EIP155Signer {
	return types.NewEIP155Signer(chainID)
}

func (e *ethAPIShim) ParseBig256(s string) (*big.Int, bool) {
	return math.ParseBig256(s)
}

func (e *ethAPIShim) S256(x *big.Int) *big.Int {
	return math.S256(x)
}

func (e *ethAPIShim) GenerateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func (e *ethAPIShim) PubkeyToAddress(p ecdsa.PublicKey) Address {
	return crypto.PubkeyToAddress(p)
}

// FromECDSA exports a private key into a binary dump
func (e *ethAPIShim) FromECDSA(priv *ecdsa.PrivateKey) []byte {
	return crypto.FromECDSA(priv)
}

// HexToECDSA parses a secp256k1 private key
func (e *ethAPIShim) HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	return crypto.HexToECDSA(hexkey)
}

// NewStream creates a new decoding stream reading from r
func (e *ethAPIShim) NewStream(r io.Reader, inputLimit uint64) *rlp.Stream {
	return rlp.NewStream(r, inputLimit)
}

// SignTx signs the transaction using the given signer and private key
func (e *ethAPIShim) SignTx(tx *types.Transaction, s types.Signer, prv *ecdsa.PrivateKey) (*types.Transaction, error) {
	return types.SignTx(tx, s, prv)
}
