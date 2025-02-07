package panictests

import (
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/wasm-vm/arwen"
	mock "github.com/ElrondNetwork/wasm-vm/mock/context"
	test "github.com/ElrondNetwork/wasm-vm/testcommon"
	"github.com/stretchr/testify/require"
)

const increment = "increment"

func TestExecution_PanicInGoWithSilentWasmer_SIGSEGV(t *testing.T) {
	code := test.GetTestSCCode("counter", "../../../")
	blockchain := test.BlockchainHookStubForCallSigSegv(code, big.NewInt(1))
	host := test.NewTestHostBuilder(t).
		WithWasmerSIGSEGVPassthrough(true).
		WithBlockchainHook(blockchain).
		Build()
	defer func() {
		host.Reset()
	}()

	blockchain.GetStorageDataCalled = func(_ []byte, _ []byte) ([]byte, uint32, error) {
		var i *int
		i = nil

		// dereference a nil pointer
		*i = *i + 1
		return nil, 0, nil
	}

	input := test.CreateTestContractCallInputBuilder().
		WithGasProvided(1000000).
		WithFunction(increment).
		Build()

	// Ensure that no more panic
	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	_, err := host.RunSmartContractCall(input)
	require.Equal(t, err, arwen.ErrExecutionPanicked)
}

func TestExecution_PanicInGoWithSilentWasmer_SIGFPE(t *testing.T) {
	code := test.GetTestSCCode("counter", "../../../")
	blockchain := test.BlockchainHookStubForCallSigSegv(code, big.NewInt(1))
	host := test.NewTestHostBuilder(t).
		WithWasmerSIGSEGVPassthrough(true).
		WithBlockchainHook(blockchain).
		Build()
	defer func() {
		host.Reset()
	}()

	blockchain.GetStorageDataCalled = func(_ []byte, _ []byte) ([]byte, uint32, error) {
		i := 5
		j := 4
		i = i / (j - 4)
		return nil, 0, nil
	}

	input := test.CreateTestContractCallInputBuilder().
		WithGasProvided(1000000).
		WithFunction(increment).
		Build()

	// Ensure that host.RunSmartContractCall() still panics, but the panic is a
	// wrapped error.
	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	_, err := host.RunSmartContractCall(input)
	require.Equal(t, err, arwen.ErrExecutionPanicked)
}

func TestExecution_PanicInGoWithSilentWasmer_Timeout(t *testing.T) {
	code := test.GetTestSCCode("counter", "../../../")
	blockchain := test.BlockchainHookStubForCallSigSegv(code, big.NewInt(1))
	host := test.NewTestHostBuilder(t).
		WithWasmerSIGSEGVPassthrough(true).
		WithBlockchainHook(blockchain).
		Build()
	defer func() {
		host.Reset()
	}()

	blockchain.GetStorageDataCalled = func(_ []byte, _ []byte) ([]byte, uint32, error) {
		time.Sleep(2 * time.Second)
		return nil, 0, nil
	}

	input := test.CreateTestContractCallInputBuilder().
		WithGasProvided(1000000).
		WithFunction(increment).
		Build()

	// Ensure that panics are not thrown
	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	_, err := host.RunSmartContractCall(input)
	require.Equal(t, err, arwen.ErrExecutionFailedWithTimeout)
}

func TestExecution_PanicInGoWithSilentWasmer_TimeoutAndSIGSEGV(t *testing.T) {
	code := test.GetTestSCCode("counter", "../../../")
	blockchain := test.BlockchainHookStubForCallSigSegv(code, big.NewInt(1))
	host := test.NewTestHostBuilder(t).
		WithWasmerSIGSEGVPassthrough(true).
		WithBlockchainHook(blockchain).
		Build()

	defer func() {
		host.Reset()
	}()

	blockchain.GetStorageDataCalled = func(_ []byte, _ []byte) ([]byte, uint32, error) {
		var i *int
		i = nil

		// dereference a nil pointer
		time.Sleep(time.Second)
		*i = *i + 1
		return nil, 0, nil
	}

	input := test.CreateTestContractCallInputBuilder().
		WithGasProvided(1000000).
		WithFunction(increment).
		Build()

	// Ensure that panics are not thrown
	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	_, err := host.RunSmartContractCall(input)
	require.NotNil(t, err)
}

func TestExecution_MultipleHostsPanicInGoWithSilentWasmer_TimeoutAndSIGSEGV(t *testing.T) {
	numParallel := 100
	hosts := make([]arwen.VMHost, numParallel)
	blockchains := make([]*mock.BlockchainHookStub, numParallel)
	for k := 0; k < numParallel; k++ {
		code := test.GetTestSCCode("counter", "../../../")
		blockchains[k] = test.BlockchainHookStubForCallSigSegv(code, big.NewInt(1))
		hosts[k] = test.NewTestHostBuilder(t).
			WithWasmerSIGSEGVPassthrough(true).
			WithBlockchainHook(blockchains[k]).
			Build()
		blockchains[k].GetStorageDataCalled = func(_ []byte, _ []byte) ([]byte, uint32, error) {
			var i *int
			i = nil

			// dereference a nil pointer
			time.Sleep(time.Second)
			*i = *i + 1
			return nil, 0, nil
		}
	}

	defer func() {
		for _, vm := range hosts {
			vm.Reset()
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(numParallel)
	for k := 0; k < numParallel; k++ {
		go func(idx int) {
			input := test.CreateTestContractCallInputBuilder().
				WithGasProvided(1000000).
				WithFunction(increment).
				Build()
			// Ensure that no more panic
			defer func() {
				r := recover()
				require.Nil(t, r)
			}()

			_, err := hosts[idx].RunSmartContractCall(input)
			require.NotNil(t, err)
			wg.Done()
		}(k)
	}

	wg.Wait()
}
