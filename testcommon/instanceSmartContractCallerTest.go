package testcommon

import (
	"testing"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/config"
	contextmock "github.com/ElrondNetwork/wasm-vm/mock/context"
)

// InstanceTestSmartContract represents the config data for the smart contract instance to be tested
type InstanceTestSmartContract struct {
	testSmartContract
	code []byte
}

// CreateInstanceContract build a contract to be used in a test creted with BuildInstanceCallTest
func CreateInstanceContract(address []byte) *InstanceTestSmartContract {
	return &InstanceTestSmartContract{
		testSmartContract: testSmartContract{
			address: address,
		},
	}
}

// WithBalance provides the balance for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithBalance(balance int64) *InstanceTestSmartContract {
	mockSC.balance = balance
	return mockSC
}

// WithConfig provides the config object for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithConfig(testConfig *TestConfig) *InstanceTestSmartContract {
	mockSC.config = testConfig
	return mockSC
}

// WithCode provides the code for the InstanceTestSmartContract
func (mockSC *InstanceTestSmartContract) WithCode(code []byte) *InstanceTestSmartContract {
	mockSC.code = code
	return mockSC
}

// InstancesTestTemplate holds the data to build a contract call test
type InstancesTestTemplate struct {
	testTemplateConfig
	contracts          []*InstanceTestSmartContract
	gasSchedule        config.GasScheduleMap
	setup              func(arwen.VMHost, *contextmock.BlockchainHookStub)
	assertResults      func(arwen.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)
	host               arwen.VMHost
	blockchainHookStub *contextmock.BlockchainHookStub
}

// BuildInstanceCallTest starts the building process for a contract call test
func BuildInstanceCallTest(tb testing.TB) *InstancesTestTemplate {
	return &InstancesTestTemplate{
		testTemplateConfig: testTemplateConfig{
			tb:                       tb,
			useMocks:                 false,
			wasmerSIGSEGVPassthrough: false,
		},
		setup: func(arwen.VMHost, *contextmock.BlockchainHookStub) {},
	}
}

// WithContracts provides the contracts to be used by the contract call test
func (callerTest *InstancesTestTemplate) WithContracts(usedContracts ...*InstanceTestSmartContract) *InstancesTestTemplate {
	callerTest.contracts = usedContracts
	return callerTest
}

// WithInput provides the ContractCallInput to be used by the contract call test
func (callerTest *InstancesTestTemplate) WithInput(input *vmcommon.ContractCallInput) *InstancesTestTemplate {
	callerTest.input = input
	return callerTest
}

// WithSetup provides the setup function to be used by the contract call test
func (callerTest *InstancesTestTemplate) WithSetup(setup func(arwen.VMHost, *contextmock.BlockchainHookStub)) *InstancesTestTemplate {
	callerTest.setup = setup
	return callerTest
}

// WithGasSchedule provides gas schedule for the test
func (callerTest *InstancesTestTemplate) WithGasSchedule(gasSchedule config.GasScheduleMap) *InstancesTestTemplate {
	callerTest.gasSchedule = gasSchedule
	return callerTest
}

// WithWasmerSIGSEGVPassthrough sets the wasmerSIGSEGVPassthrough flag
func (callerTest *InstancesTestTemplate) WithWasmerSIGSEGVPassthrough(wasmerSIGSEGVPassthrough bool) *InstancesTestTemplate {
	callerTest.wasmerSIGSEGVPassthrough = wasmerSIGSEGVPassthrough
	return callerTest
}

// AndAssertResults starts the test and asserts the results
func (callerTest *InstancesTestTemplate) AndAssertResults(assertResults func(arwen.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)) {
	callerTest.assertResults = assertResults
	runTestWithInstances(callerTest, true)
}

// AndAssertResultsWithoutReset starts the test and asserts the results
func (callerTest *InstancesTestTemplate) AndAssertResultsWithoutReset(assertResults func(arwen.VMHost, *contextmock.BlockchainHookStub, *VMOutputVerifier)) {
	callerTest.assertResults = assertResults
	runTestWithInstances(callerTest, false)
}

func runTestWithInstances(callerTest *InstancesTestTemplate, reset bool) {
	if callerTest.host == nil {
		callerTest.blockchainHookStub = BlockchainHookStubForContracts(callerTest.contracts)
		callerTest.host = NewTestHostBuilder(callerTest.tb).
			WithBlockchainHook(callerTest.blockchainHookStub).
			WithGasSchedule(callerTest.gasSchedule).
			WithWasmerSIGSEGVPassthrough(callerTest.wasmerSIGSEGVPassthrough).
			Build()
		callerTest.setup(callerTest.host, callerTest.blockchainHookStub)
	}

	defer func() {
		if reset {
			callerTest.host.Reset()
		}
	}()

	vmOutput, err := callerTest.host.RunSmartContractCall(callerTest.input)

	if callerTest.assertResults != nil {
		allErrors := callerTest.host.Runtime().GetAllErrors()
		verify := NewVMOutputVerifierWithAllErrors(callerTest.tb, vmOutput, err, allErrors)
		callerTest.assertResults(callerTest.host, callerTest.blockchainHookStub, verify)
	}
}
