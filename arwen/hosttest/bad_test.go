package hosttest

import (
	"testing"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/executor"
	contextmock "github.com/ElrondNetwork/wasm-vm/mock/context"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
)

func TestBadContract_NoPanic_Memoryfault(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("memoryFault").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrorAndInfo(arwen.ErrExecutionFailed.Error(), "memoryFault")
		})
}

func TestBadContract_NoPanic_DivideByZero(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("divideByZero").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				Ok()
		})
}

func TestBadContract_NoPanic_BadGetOwner1(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badGetOwner1").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors(arwen.ErrBadBounds.Error())
		})
}

func TestBadContract_NoPanic_BadBigIntStorageStore1(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badBigIntStorageStore1").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				Ok()
		})
}

func TestBadContract_NoPanic_BadWriteLog1(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badWriteLog1").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors("negative length")
		})
}

func TestBadContract_NoPanic_BadWriteLog2(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badWriteLog2").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors("negative length")
		})
}

func TestBadContract_NoPanic_BadWriteLog3(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badWriteLog3").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				Ok()
		})
}

func TestBadContract_NoPanic_BadWriteLog4(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badWriteLog4").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors("mem load: bad bounds")
		})
}

func TestBadContract_NoPanic_BadGetBlockHash1(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badGetBlockHash1").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors(arwen.ErrExecutionFailed.Error())
		})
}

func TestBadContract_NoPanic_BadGetBlockHash2(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badGetBlockHash2").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				Ok()
		})
}

func TestBadContract_NoPanic_BadGetBlockHash3(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("badGetBlockHash3").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors(arwen.ErrExecutionFailed.Error())
		})
}

func TestBadContract_NoPanic_BadRecursive(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-misc", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(10000000).
			WithFunction("badRecursive").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				ExecutionFailed().
				HasRuntimeErrors(arwen.ErrExecutionFailed.Error())
		})
}

func TestBadContract_NoPanic_NonExistingFunction(t *testing.T) {
	test.BuildInstanceCallTest(t).
		WithContracts(
			test.CreateInstanceContract(test.ParentAddress).
				WithCode(test.GetTestSCCode("bad-empty", "../../")).
				WithBalance(1000)).
		WithInput(test.CreateTestContractCallInputBuilder().
			WithRecipientAddr(test.ParentAddress).
			WithGasProvided(test.GasProvided).
			WithFunction("thisDoesNotExist").
			Build()).
		WithWasmerSIGSEGVPassthrough(false).
		AndAssertResults(func(host arwen.VMHost, stubBlockchainHook *contextmock.BlockchainHookStub, verify *test.VMOutputVerifier) {
			verify.
				FunctionNotFound().
				HasRuntimeErrorAndInfo(executor.ErrInvalidFunction.Error(), "thisDoesNotExist")
		})
}
