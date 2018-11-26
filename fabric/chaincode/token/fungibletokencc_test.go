package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
)

const ccName = "tokenChaincode"

const owner = "000000"

const supply = 10000

func TestInit(t *testing.T) {
	stub := newMockStub()
	r := initMock(stub)
	assert.Equal(t, shim.OK, int(r.Status), r.Message)

	// Check initial token state
	token, err := readToken(stub)
	assert.NoError(t, err)
	assert.Equal(t, *token, Token{Symbol: "FUSD", Name: "Fabric USD", Decimals: 0, Supply: supply})
}

func TestInvoke(t *testing.T) {
	stub := newMockStub()
	r := initMock(stub)
	assert.Equal(t, shim.OK, int(r.Status), r.Message)

	r = stub.MockInvokeWithSignedProposal("1", byteArray("TokenSupply"), &pb.SignedProposal{})
	assert.Equal(t, shim.OK, int(r.Status), r.Message)
	assert.Equal(t, strconv.Itoa(supply), string(r.Payload))

	r = stub.MockInvokeWithSignedProposal("1", byteArray("BalanceOf", owner), &pb.SignedProposal{})
	assert.Equal(t, shim.OK, int(r.Status), r.Message)
	assert.Equal(t, strconv.Itoa(supply), string(r.Payload))

	r = stub.MockInvokeWithSignedProposal("1", byteArray("Transfer", "dileban", "100"), &pb.SignedProposal{})
	assert.Equal(t, shim.OK, int(r.Status), r.Message)
	assert.Nil(t, r.Payload)
}

func newMockStub() *shim.MockStub {
	return shim.NewMockStub(ccName, new(TokenChaincode))
}

func initMock(stub *shim.MockStub) pb.Response {
	return stub.MockInit("init", byteArray("FUSD", "Fabric USD", "10000", owner))
}

func readToken(stub *shim.MockStub) (*Token, error) {
	var t Token
	var b []byte
	var ok bool
	if b, ok = stub.State["token"]; !ok {
		return nil, fmt.Errorf("Error reading token")
	}
	if err := json.Unmarshal(b, &t); err != nil {
		return &t, fmt.Errorf("Error unmarshaling token")
	}
	return &t, nil
}

func readBalance(stub *shim.MockStub, address string) (*Balance, error) {
	var bal Balance
	var b []byte
	var ok bool
	if b, ok = stub.State[address]; !ok {
		return nil, fmt.Errorf("Error reading balance")
	}
	if err := json.Unmarshal(b, &bal); err != nil {
		return nil, fmt.Errorf("Error unmarkshalling balance")
	}
	return &bal, nil
}

func byteArray(s ...string) [][]byte {
	args := make([][]byte, len(s))
	for i, v := range s {
		args[i] = []byte(v)
	}
	return args
}
