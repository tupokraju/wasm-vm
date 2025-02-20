package host

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/arwen/contexts"
	"github.com/ElrondNetwork/wasm-vm/arwen/elrondapi"
	"github.com/ElrondNetwork/wasm-vm/config"
	"github.com/ElrondNetwork/wasm-vm/crypto"
	"github.com/ElrondNetwork/wasm-vm/crypto/factory"
	"github.com/ElrondNetwork/wasm-vm/executor"
	"github.com/ElrondNetwork/wasm-vm/wasmer"
)

var log = logger.GetOrCreate("arwen/host")
var logGasTrace = logger.GetOrCreate("gasTrace")

// MaximumRuntimeInstanceStackSize specifies the maximum number of allowed Wasmer
// instances on the InstanceStack of the RuntimeContext
var MaximumRuntimeInstanceStackSize = uint64(10)

var _ arwen.VMHost = (*vmHost)(nil)

const minExecutionTimeout = time.Second
const internalVMErrors = "internalVMErrors"

var defaultVMExecutorFactory executor.ExecutorAbstractFactory = wasmer.ExecutorFactory()

// vmHost implements HostContext interface.
type vmHost struct {
	cryptoHook       crypto.VMCrypto
	mutExecution     sync.RWMutex
	closingInstance  bool
	executionTimeout time.Duration

	ethInput []byte

	blockchainContext   arwen.BlockchainContext
	runtimeContext      arwen.RuntimeContext
	asyncContext        arwen.AsyncContext
	outputContext       arwen.OutputContext
	meteringContext     arwen.MeteringContext
	storageContext      arwen.StorageContext
	managedTypesContext arwen.ManagedTypesContext

	gasSchedule          config.GasScheduleMap
	builtInFuncContainer vmcommon.BuiltInFunctionContainer
	esdtTransferParser   vmcommon.ESDTTransferParser
	callArgsParser       arwen.CallArgsParser
	enableEpochsHandler  vmcommon.EnableEpochsHandler
	activationEpochMap   map[uint32]struct{}
}

// NewArwenVM creates a new Arwen vmHost
func NewArwenVM(
	blockChainHook vmcommon.BlockchainHook,
	hostParameters *arwen.VMHostParameters,
) (arwen.VMHost, error) {

	if check.IfNil(blockChainHook) {
		return nil, arwen.ErrNilBlockChainHook
	}
	if hostParameters == nil {
		return nil, arwen.ErrNilHostParameters
	}
	if check.IfNil(hostParameters.ESDTTransferParser) {
		return nil, arwen.ErrNilESDTTransferParser
	}
	if check.IfNil(hostParameters.BuiltInFuncContainer) {
		return nil, arwen.ErrNilBuiltInFunctionsContainer
	}
	if check.IfNil(hostParameters.EpochNotifier) {
		return nil, arwen.ErrNilEpochNotifier
	}
	if check.IfNil(hostParameters.EnableEpochsHandler) {
		return nil, arwen.ErrNilEnableEpochsHandler
	}

	cryptoHook := factory.NewVMCrypto()
	host := &vmHost{
		cryptoHook:           cryptoHook,
		meteringContext:      nil,
		runtimeContext:       nil,
		asyncContext:         nil,
		blockchainContext:    nil,
		storageContext:       nil,
		managedTypesContext:  nil,
		gasSchedule:          hostParameters.GasSchedule,
		builtInFuncContainer: hostParameters.BuiltInFuncContainer,
		esdtTransferParser:   hostParameters.ESDTTransferParser,
		callArgsParser:       parsers.NewCallArgsParser(),
		executionTimeout:     minExecutionTimeout,
		enableEpochsHandler:  hostParameters.EnableEpochsHandler,
	}
	newExecutionTimeout := time.Duration(hostParameters.TimeOutForSCExecutionInMilliseconds) * time.Millisecond
	if newExecutionTimeout > minExecutionTimeout {
		host.executionTimeout = newExecutionTimeout
	}

	var err error
	host.blockchainContext, err = contexts.NewBlockchainContext(host, blockChainHook)
	if err != nil {
		return nil, err
	}
	vmExecutor, err := host.createExecutor(hostParameters)
	if err != nil {
		return nil, err
	}
	host.runtimeContext, err = contexts.NewRuntimeContext(
		host,
		hostParameters.VMType,
		host.builtInFuncContainer,
		vmExecutor,
	)
	if err != nil {
		return nil, err
	}

	host.meteringContext, err = contexts.NewMeteringContext(host, hostParameters.GasSchedule, hostParameters.BlockGasLimit)
	if err != nil {
		return nil, err
	}

	host.outputContext, err = contexts.NewOutputContext(host)
	if err != nil {
		return nil, err
	}

	host.storageContext, err = contexts.NewStorageContext(
		host,
		blockChainHook,
		hostParameters.ElrondProtectedKeyPrefix,
	)
	if err != nil {
		return nil, err
	}

	host.asyncContext, err = contexts.NewAsyncContext(host, host.callArgsParser, host.esdtTransferParser, &marshal.GogoProtoMarshalizer{})
	if err != nil {
		return nil, err
	}

	host.managedTypesContext, err = contexts.NewManagedTypesContext(host)
	if err != nil {
		return nil, err
	}

	host.runtimeContext.SetMaxInstanceStackSize(MaximumRuntimeInstanceStackSize)

	host.initContexts()
	hostParameters.EpochNotifier.RegisterNotifyHandler(host)

	return host, nil
}

// Creates a new executor instance. Should only be called once per VM host instantiation.
func (host *vmHost) createExecutor(hostParameters *arwen.VMHostParameters) (executor.Executor, error) {
	vmHooks := elrondapi.NewElrondApi(host)
	gasCostConfig, err := config.CreateGasConfig(host.gasSchedule)
	if err != nil {
		return nil, err
	}
	vmExecutorFactory := defaultVMExecutorFactory
	if hostParameters.OverrideVMExecutor != nil {
		vmExecutorFactory = hostParameters.OverrideVMExecutor
	}
	vmExecutorFactoryArgs := executor.ExecutorFactoryArgs{
		VMHooks:                  vmHooks,
		OpcodeCosts:              gasCostConfig.WASMOpcodeCost,
		RkyvSerializationEnabled: true,
		WasmerSIGSEGVPassthrough: hostParameters.WasmerSIGSEGVPassthrough,
	}
	return vmExecutorFactory.CreateExecutor(vmExecutorFactoryArgs)
}

func createActivationMap(hostParameters *arwen.VMHostParameters) map[uint32]struct{} {
	activationMap := make(map[uint32]struct{})

	activationMap[hostParameters.EnableEpochsHandler.CheckExecuteReadOnlyEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.DisableExecByCallerEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.RefactorContextEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.FixFailExecutionOnErrorEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.ManagedCryptoAPIEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.CreateNFTThroughExecByCallerEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.FixOOGReturnCodeEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.MultiESDTTransferAsyncCallBackEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.RemoveNonUpdatedStorageEnableEpoch()] = struct{}{}
	activationMap[hostParameters.EnableEpochsHandler.StorageAPICostOptimizationEnableEpoch()] = struct{}{}

	return activationMap
}

// GetVersion returns the Arwen version string
func (host *vmHost) GetVersion() string {
	return arwen.ArwenVersion
}

// Crypto returns the VMCrypto instance of the host
func (host *vmHost) Crypto() crypto.VMCrypto {
	return host.cryptoHook
}

// Blockchain returns the BlockchainContext instance of the host
func (host *vmHost) Blockchain() arwen.BlockchainContext {
	return host.blockchainContext
}

// Runtime returns the RuntimeContext instance of the host
func (host *vmHost) Runtime() arwen.RuntimeContext {
	return host.runtimeContext
}

// Output returns the OutputContext instance of the host
func (host *vmHost) Output() arwen.OutputContext {
	return host.outputContext
}

// Metering returns the MeteringContext instance of the host
func (host *vmHost) Metering() arwen.MeteringContext {
	return host.meteringContext
}

// Async returns the AsyncContext instance of the host
func (host *vmHost) Async() arwen.AsyncContext {
	return host.asyncContext
}

// Storage returns the StorageContext instance of the host
func (host *vmHost) Storage() arwen.StorageContext {
	return host.storageContext
}

// EnableEpochsHandler returns the enableEpochsHandler instance of the host
func (host *vmHost) EnableEpochsHandler() vmcommon.EnableEpochsHandler {
	return host.enableEpochsHandler
}

// ManagedTypes returns the ManagedTypeContext instance of the host
func (host *vmHost) ManagedTypes() arwen.ManagedTypesContext {
	return host.managedTypesContext
}

// GetContexts returns the main contexts of the host
func (host *vmHost) GetContexts() (
	arwen.ManagedTypesContext,
	arwen.BlockchainContext,
	arwen.MeteringContext,
	arwen.OutputContext,
	arwen.RuntimeContext,
	arwen.AsyncContext,
	arwen.StorageContext,
) {
	return host.managedTypesContext,
		host.blockchainContext,
		host.meteringContext,
		host.outputContext,
		host.runtimeContext,
		host.asyncContext,
		host.storageContext
}

// InitState resets the contexts of the host and reconfigures its flags
func (host *vmHost) InitState() {
	host.initContexts()
}

func (host *vmHost) close() {
	host.runtimeContext.ClearWarmInstanceCache()
}

// Close will close all underlying processes
func (host *vmHost) Close() error {
	host.mutExecution.Lock()
	host.close()
	host.closingInstance = true
	host.mutExecution.Unlock()

	return nil
}

// Reset is a function which closes the VM and resets the closingInstance variable
func (host *vmHost) Reset() {
	host.mutExecution.Lock()
	host.close()
	// keep closingInstance flag to false
	host.mutExecution.Unlock()
}

func (host *vmHost) initContexts() {
	host.ClearContextStateStack()
	host.managedTypesContext.InitState()
	host.outputContext.InitState()
	host.meteringContext.InitState()
	host.runtimeContext.InitState()
	host.asyncContext.InitState()
	host.storageContext.InitState()
	host.blockchainContext.InitState()
	host.ethInput = nil
}

// ClearContextStateStack cleans the state stacks of all the contexts of the host
func (host *vmHost) ClearContextStateStack() {
	host.managedTypesContext.ClearStateStack()
	host.outputContext.ClearStateStack()
	host.meteringContext.ClearStateStack()
	host.runtimeContext.ClearStateStack()
	host.asyncContext.ClearStateStack()
	host.storageContext.ClearStateStack()
	host.blockchainContext.ClearStateStack()
}

// GasScheduleChange applies a new gas schedule to the host
func (host *vmHost) GasScheduleChange(newGasSchedule config.GasScheduleMap) {
	host.mutExecution.Lock()
	defer host.mutExecution.Unlock()

	host.gasSchedule = newGasSchedule
	gasCostConfig, err := config.CreateGasConfig(newGasSchedule)
	if err != nil {
		log.Error("cannot apply new gas config", "err", err)
		return
	}

	host.runtimeContext.GetVMExecutor().SetOpcodeCosts(gasCostConfig.WASMOpcodeCost)

	host.meteringContext.SetGasSchedule(newGasSchedule)
	host.runtimeContext.ClearWarmInstanceCache()
}

// GetGasScheduleMap returns the currently stored gas schedule
func (host *vmHost) GetGasScheduleMap() config.GasScheduleMap {
	return host.gasSchedule
}

// RunSmartContractCreate executes the deployment of a new contract
func (host *vmHost) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (vmOutput *vmcommon.VMOutput, err error) {
	host.mutExecution.RLock()
	defer host.mutExecution.RUnlock()

	if host.closingInstance {
		return nil, arwen.ErrVMIsClosing
	}

	host.setGasTracerEnabledIfLogIsTrace()
	ctx, cancel := context.WithTimeout(context.Background(), host.executionTimeout)
	defer cancel()

	log.Trace("RunSmartContractCreate begin",
		"len(code)", len(input.ContractCode),
		"metadata", input.ContractCodeMetadata,
		"gasProvided", input.GasProvided,
		"gasLocked", input.GasLocked)

	done := make(chan struct{})
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Error("VM execution panicked", "error", r, "stack", "\n"+string(debug.Stack()))
				err = arwen.ErrExecutionPanicked
				host.Runtime().CleanInstance()
			}

			close(done)
		}()

		vmOutput = host.doRunSmartContractCreate(input)
		logsFromErrors := host.createLogEntryFromErrors(input.CallerAddr, input.CallerAddr, "_init")
		if logsFromErrors != nil {
			vmOutput.Logs = append(vmOutput.Logs, logsFromErrors)
		}

		log.Trace("RunSmartContractCreate end",
			"returnCode", vmOutput.ReturnCode,
			"returnMessage", vmOutput.ReturnMessage,
			"gasRemaining", vmOutput.GasRemaining)
		host.logFromGasTracer("init")
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		host.Runtime().FailExecution(arwen.ErrExecutionFailedWithTimeout)
		<-done
		err = arwen.ErrExecutionFailedWithTimeout
	}

	return
}

// RunSmartContractCall executes the call of an existing contract
func (host *vmHost) RunSmartContractCall(input *vmcommon.ContractCallInput) (vmOutput *vmcommon.VMOutput, err error) {
	host.mutExecution.RLock()
	defer host.mutExecution.RUnlock()

	if host.closingInstance {
		return nil, arwen.ErrVMIsClosing
	}

	host.setGasTracerEnabledIfLogIsTrace()
	ctx, cancel := context.WithTimeout(context.Background(), host.executionTimeout)
	defer cancel()

	log.Trace("RunSmartContractCall begin",
		"function", input.Function,
		"gasProvided", input.GasProvided,
		"gasLocked", input.GasLocked)

	done := make(chan struct{})
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				log.Error("VM execution panicked", "error", r, "stack", "\n"+string(debug.Stack()))
				err = arwen.ErrExecutionPanicked
				host.Runtime().CleanInstance()
			}

			close(done)
		}()

		switch input.Function {
		case arwen.UpgradeFunctionName:
			vmOutput = host.doRunSmartContractUpgrade(input)
		case arwen.DeleteFunctionName:
			vmOutput = host.doRunSmartContractDelete(input)
		default:
			vmOutput = host.doRunSmartContractCall(input)
		}

		logsFromErrors := host.createLogEntryFromErrors(input.CallerAddr, input.RecipientAddr, input.Function)
		if logsFromErrors != nil {
			vmOutput.Logs = append(vmOutput.Logs, logsFromErrors)
		}

		log.Trace("RunSmartContractCall end",
			"function", input.Function,
			"returnCode", vmOutput.ReturnCode,
			"returnMessage", vmOutput.ReturnMessage,
			"gasRemaining", vmOutput.GasRemaining)
		host.logFromGasTracer(input.Function)
	}()

	select {
	case <-done:
		// Normal termination.
		return
	case <-ctx.Done():
		// Terminated due to timeout. The VM sets the `ExecutionFailed` breakpoint
		// in Wasmer. Also, the VM must wait for Wasmer to reach the end of a WASM
		// basic block in order to close the WASM instance cleanly. This is done by
		// reading the `done` channel once more, awaiting the call to `close(done)`
		// from above.
		host.Runtime().FailExecution(arwen.ErrExecutionFailedWithTimeout)
		<-done
		err = arwen.ErrExecutionFailedWithTimeout
	}

	return
}

func (host *vmHost) createLogEntryFromErrors(sndAddress, rcvAddress []byte, function string) *vmcommon.LogEntry {
	formattedErrors := host.runtimeContext.GetAllErrors()
	if formattedErrors == nil {
		return nil
	}

	logFromError := &vmcommon.LogEntry{
		Identifier: []byte(internalVMErrors),
		Address:    sndAddress,
		Topics:     [][]byte{rcvAddress, []byte(function)},
		Data:       []byte(formattedErrors.Error()),
	}

	return logFromError
}

// AreInSameShard returns true if the provided addresses are part of the same shard
func (host *vmHost) AreInSameShard(leftAddress []byte, rightAddress []byte) bool {
	blockchain := host.Blockchain()
	leftShard := blockchain.GetShardOfAddress(leftAddress)
	rightShard := blockchain.GetShardOfAddress(rightAddress)

	return leftShard == rightShard
}

// IsInterfaceNil returns true if there is no value under the interface
func (host *vmHost) IsInterfaceNil() bool {
	return host == nil
}

// SetRuntimeContext sets the runtimeContext for this host, used in tests
func (host *vmHost) SetRuntimeContext(runtime arwen.RuntimeContext) {
	host.runtimeContext = runtime
}

// GetRuntimeErrors obtains the cumultated error object after running the SC
func (host *vmHost) GetRuntimeErrors() error {
	if host.runtimeContext != nil {
		return host.runtimeContext.GetAllErrors()
	}
	return nil
}

// SetBuiltInFunctionsContainer sets the built in function container - only for testing
func (host *vmHost) SetBuiltInFunctionsContainer(builtInFuncs vmcommon.BuiltInFunctionContainer) {
	if check.IfNil(builtInFuncs) {
		return
	}
	host.builtInFuncContainer = builtInFuncs
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (host *vmHost) EpochConfirmed(epoch uint32, _ uint64) {
	_, ok := host.activationEpochMap[epoch]
	if ok {
		host.Runtime().ClearWarmInstanceCache()
		host.Blockchain().ClearCompiledCodes()
	}
}

// FixOOGReturnCodeEnabled returns true if the corresponding flag is set
func (host *vmHost) FixOOGReturnCodeEnabled() bool {
	return host.enableEpochsHandler.IsFixOOGReturnCodeFlagEnabled()
}

// FixFailExecutionEnabled returns true if the corresponding flag is set
func (host *vmHost) FixFailExecutionEnabled() bool {
	return host.enableEpochsHandler.IsFailExecutionOnEveryAPIErrorFlagEnabled()
}

// CreateNFTOnExecByCallerEnabled returns true if the corresponding flag is set
func (host *vmHost) CreateNFTOnExecByCallerEnabled() bool {
	return host.enableEpochsHandler.IsCreateNFTThroughExecByCallerFlagEnabled()
}

// DisableExecByCaller returns true if the corresponding flag is set
func (host *vmHost) DisableExecByCaller() bool {
	return host.enableEpochsHandler.IsDisableExecByCallerFlagEnabled()
}

// CheckExecuteReadOnly returns true if the corresponding flag is set
func (host *vmHost) CheckExecuteReadOnly() bool {
	return host.enableEpochsHandler.IsCheckExecuteOnReadOnlyFlagEnabled()
}

func (host *vmHost) setGasTracerEnabledIfLogIsTrace() {
	host.Metering().SetGasTracing(false)
	if logGasTrace.GetLevel() == logger.LogTrace {
		host.Metering().SetGasTracing(true)
	}
}

func (host *vmHost) logFromGasTracer(functionName string) {
	if logGasTrace.GetLevel() == logger.LogTrace {
		scGasTrace := host.meteringContext.GetGasTrace()
		totalGasUsedByAPIs := 0
		for scAddress, gasTrace := range scGasTrace {
			logGasTrace.Trace("Gas Trace for", "SC Address", scAddress, "function", functionName)
			for apiName, value := range gasTrace {
				totalGasUsed := uint64(0)
				for _, usedGas := range value {
					totalGasUsed += usedGas
				}
				logGasTrace.Trace("Gas Trace for", "apiName", apiName, "totalGasUsed", totalGasUsed, "numberOfCalls", len(value))
				totalGasUsedByAPIs += int(totalGasUsed)
			}
			logGasTrace.Trace("Gas Trace for", "TotalGasUsedByAPIs", totalGasUsedByAPIs)
		}
	}
}
