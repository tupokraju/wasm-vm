package elrondapi

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	"github.com/ElrondNetwork/wasm-vm/math"
)

const (
	getSCAddressName                 = "getSCAddress"
	getOwnerAddressName              = "getOwnerAddress"
	getShardOfAddressName            = "getShardOfAddress"
	isSmartContractName              = "isSmartContract"
	getExternalBalanceName           = "getExternalBalance"
	blockHashName                    = "blockHash"
	transferValueName                = "transferValue"
	transferESDTExecuteName          = "transferESDTExecute"
	transferESDTNFTExecuteName       = "transferESDTNFTExecute"
	multiTransferESDTNFTExecuteName  = "multiTransferESDTNFTExecute"
	transferValueExecuteName         = "transferValueExecute"
	createAsyncCallName              = "createAsyncCall"
	setAsyncGroupCallbackName        = "setAsyncGroupCallback"
	setAsyncContextCallbackName      = "setAsyncContextCallback"
	getArgumentLengthName            = "getArgumentLength"
	getArgumentName                  = "getArgument"
	getFunctionName                  = "getFunction"
	getNumArgumentsName              = "getNumArguments"
	storageStoreName                 = "storageStore"
	storageLoadLengthName            = "storageLoadLength"
	storageLoadName                  = "storageLoad"
	storageLoadFromAddressName       = "storageLoadFromAddress"
	getCallerName                    = "getCaller"
	checkNoPaymentName               = "checkNoPayment"
	callValueName                    = "callValue"
	getESDTValueName                 = "getESDTValue"
	getESDTTokenNameName             = "getESDTTokenName"
	getESDTTokenNonceName            = "getESDTTokenNonce"
	getESDTTokenTypeName             = "getESDTTokenType"
	getCallValueTokenNameName        = "getCallValueTokenName"
	getESDTValueByIndexName          = "getESDTValueByIndex"
	getESDTTokenNameByIndexName      = "getESDTTokenNameByIndex"
	getESDTTokenNonceByIndexName     = "getESDTTokenNonceByIndex"
	getESDTTokenTypeByIndexName      = "getESDTTokenTypeByIndex"
	getCallValueTokenNameByIndexName = "getCallValueTokenNameByIndex"
	getNumESDTTransfersName          = "getNumESDTTransfers"
	getCurrentESDTNFTNonceName       = "getCurrentESDTNFTNonce"
	writeLogName                     = "writeLog"
	writeEventLogName                = "writeEventLog"
	returnDataName                   = "returnData"
	signalErrorName                  = "signalError"
	getGasLeftName                   = "getGasLeft"
	getESDTBalanceName               = "getESDTBalance"
	getESDTNFTNameLengthName         = "getESDTNFTNameLength"
	getESDTNFTAttributeLengthName    = "getESDTNFTAttributeLength"
	getESDTNFTURILengthName          = "getESDTNFTURILength"
	getESDTTokenDataName             = "getESDTTokenData"
	getESDTLocalRolesName            = "getESDTLocalRoles"
	validateTokenIdentifierName      = "validateTokenIdentifier"
	executeOnDestContextName         = "executeOnDestContext"
	executeOnSameContextName         = "executeOnSameContext"
	executeReadOnlyName              = "executeReadOnly"
	createContractName               = "createContract"
	deployFromSourceContractName     = "deployFromSourceContract"
	upgradeContractName              = "upgradeContract"
	upgradeFromSourceContractName    = "upgradeFromSourceContract"
	deleteContractName               = "deleteContract"
	asyncCallName                    = "asyncCall"
	getNumReturnDataName             = "getNumReturnData"
	getReturnDataSizeName            = "getReturnDataSize"
	getReturnDataName                = "getReturnData"
	cleanReturnDataName              = "cleanReturnData"
	deleteFromReturnDataName         = "deleteFromReturnData"
	setStorageLockName               = "setStorageLock"
	getStorageLockName               = "getStorageLock"
	isStorageLockedName              = "isStorageLocked"
	clearStorageLockName             = "clearStorageLock"
	getBlockTimestampName            = "getBlockTimestamp"
	getBlockNonceName                = "getBlockNonce"
	getBlockRoundName                = "getBlockRound"
	getBlockEpochName                = "getBlockEpoch"
	getBlockRandomSeedName           = "getBlockRandomSeed"
	getStateRootHashName             = "getStateRootHash"
	getPrevBlockTimestampName        = "getPrevBlockTimestamp"
	getPrevBlockNonceName            = "getPrevBlockNonce"
	getPrevBlockRoundName            = "getPrevBlockRound"
	getPrevBlockEpochName            = "getPrevBlockEpoch"
	getPrevBlockRandomSeedName       = "getPrevBlockRandomSeed"
	getOriginalTxHashName            = "getOriginalTxHash"
	getCurrentTxHashName             = "getCurrentTxHash"
	getPrevTxHashName                = "getPrevTxHash"
)

var logEEI = logger.GetOrCreate("arwen/eei")

func getESDTTransferFromInputFailIfWrongIndex(host arwen.VMHost, index int32) *vmcommon.ESDTTransfer {
	esdtTransfers := host.Runtime().GetVMInput().ESDTTransfers
	if int32(len(esdtTransfers))-1 < index || index < 0 {
		WithFaultAndHost(host, arwen.ErrInvalidTokenIndex, host.Runtime().ElrondAPIErrorShouldFailExecution())
		return nil
	}
	return esdtTransfers[index]
}

func failIfMoreThanOneESDTTransfer(context *ElrondApi) bool {
	runtime := context.GetRuntimeContext()
	if len(runtime.GetVMInput().ESDTTransfers) > 1 {
		return context.WithFault(arwen.ErrTooManyESDTTransfers, true)
	}
	return false
}

// GetGasLeft VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetGasLeft() int64 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetGasLeft
	metering.UseGasAndAddTracedGas(getGasLeftName, gasToUse)

	return int64(metering.GasLeft())
}

// GetSCAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetSCAddress(resultOffset int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetSCAddress
	metering.UseGasAndAddTracedGas(getSCAddressName, gasToUse)

	owner := runtime.GetContextAddress()
	err := runtime.MemStore(resultOffset, owner)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// GetOwnerAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetOwnerAddress(resultOffset int32) {
	blockchain := context.GetBlockchainContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetOwnerAddress
	metering.UseGasAndAddTracedGas(getOwnerAddressName, gasToUse)

	owner, err := blockchain.GetOwnerAddress()
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	err = runtime.MemStore(resultOffset, owner)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// GetShardOfAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetShardOfAddress(addressOffset int32) int32 {
	blockchain := context.GetBlockchainContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetShardOfAddress
	metering.UseGasAndAddTracedGas(getShardOfAddressName, gasToUse)

	address, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(blockchain.GetShardOfAddress(address))
}

// IsSmartContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) IsSmartContract(addressOffset int32) int32 {
	blockchain := context.GetBlockchainContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.IsSmartContract
	metering.UseGasAndAddTracedGas(isSmartContractName, gasToUse)

	address, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	isSmartContract := blockchain.IsSmartContract(address)

	return int32(arwen.BooleanToInt(isSmartContract))
}

// SignalError VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) SignalError(messageOffset int32, messageLength int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(signalErrorName)

	gasToUse := metering.GasSchedule().ElrondAPICost.SignalError
	gasToUse += metering.GasSchedule().BaseOperationCost.PersistPerByte * uint64(messageLength)

	err := metering.UseGasBounded(gasToUse)
	if err != nil {
		_ = context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
		return
	}

	message, err := runtime.MemLoad(messageOffset, messageLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
	runtime.SignalUserError(string(message))
}

// GetExternalBalance VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetExternalBalance(addressOffset int32, resultOffset int32) {
	blockchain := context.GetBlockchainContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetExternalBalance
	metering.UseGasAndAddTracedGas(getExternalBalanceName, gasToUse)

	address, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	balance := blockchain.GetBalance(address)

	err = runtime.MemStore(resultOffset, balance)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// GetBlockHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockHash(nonce int64, resultOffset int32) int32 {
	blockchain := context.GetBlockchainContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockHash
	metering.UseGasAndAddTracedGas(blockHashName, gasToUse)

	hash := blockchain.BlockHash(uint64(nonce))
	err := runtime.MemStore(resultOffset, hash)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

func getESDTDataFromBlockchainHook(
	context *ElrondApi,
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
) (*esdt.ESDigitalToken, error) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	blockchain := context.GetBlockchainContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetExternalBalance
	metering.UseAndTraceGas(gasToUse)

	address, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if err != nil {
		return nil, err
	}

	tokenID, err := runtime.MemLoad(tokenIDOffset, tokenIDLen)
	if err != nil {
		return nil, err
	}

	esdtToken, err := blockchain.GetESDTToken(address, tokenID, uint64(nonce))
	if err != nil {
		return nil, err
	}

	return esdtToken, nil
}

// GetESDTBalance VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTBalance(
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
	resultOffset int32,
) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTBalanceName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}
	err = runtime.MemStore(resultOffset, esdtData.Value.Bytes())
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(esdtData.Value.Bytes()))
}

// GetESDTNFTNameLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTNFTNameLength(
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTNameLengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		context.WithFault(arwen.ErrNilESDTData, runtime.ElrondAPIErrorShouldFailExecution())
		return 0
	}

	return int32(len(esdtData.TokenMetaData.Name))
}

// GetESDTNFTAttributeLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTNFTAttributeLength(
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTAttributeLengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		context.WithFault(arwen.ErrNilESDTData, runtime.ElrondAPIErrorShouldFailExecution())
		return 0
	}

	return int32(len(esdtData.TokenMetaData.Attributes))
}

// GetESDTNFTURILength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTNFTURILength(
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTNFTURILengthName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}
	if esdtData == nil || esdtData.TokenMetaData == nil {
		context.WithFault(arwen.ErrNilESDTData, runtime.ElrondAPIErrorShouldFailExecution())
		return 0
	}
	if len(esdtData.TokenMetaData.URIs) == 0 {
		return 0
	}

	return int32(len(esdtData.TokenMetaData.URIs[0]))
}

// GetESDTTokenData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenData(
	addressOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	nonce int64,
	valueHandle int32,
	propertiesOffset int32,
	hashOffset int32,
	nameOffset int32,
	attributesOffset int32,
	creatorOffset int32,
	royaltiesHandle int32,
	urisOffset int32,
) int32 {
	managedType := context.GetManagedTypesContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(getESDTTokenDataName)

	esdtData, err := getESDTDataFromBlockchainHook(context, addressOffset, tokenIDOffset, tokenIDLen, nonce)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	value := managedType.GetBigIntOrCreate(valueHandle)
	value.Set(esdtData.Value)

	err = runtime.MemStore(propertiesOffset, esdtData.Properties)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	if esdtData.TokenMetaData != nil {
		err = runtime.MemStore(hashOffset, esdtData.TokenMetaData.Hash)
		if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return -1
		}
		err = runtime.MemStore(nameOffset, esdtData.TokenMetaData.Name)
		if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return -1
		}
		err = runtime.MemStore(attributesOffset, esdtData.TokenMetaData.Attributes)
		if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return -1
		}
		err = runtime.MemStore(creatorOffset, esdtData.TokenMetaData.Creator)
		if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return -1
		}

		royalties := managedType.GetBigIntOrCreate(royaltiesHandle)
		royalties.SetUint64(uint64(esdtData.TokenMetaData.Royalties))

		if len(esdtData.TokenMetaData.URIs) > 0 {
			err = runtime.MemStore(urisOffset, esdtData.TokenMetaData.URIs[0])
			if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
				return -1
			}
		}
	}
	return int32(len(esdtData.Value.Bytes()))
}

// GetESDTLocalRoles VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTLocalRoles(tokenIdHandle int32) int64 {
	managedType := context.GetManagedTypesContext()
	runtime := context.GetRuntimeContext()
	storage := context.GetStorageContext()
	metering := context.GetMeteringContext()

	tokenID, err := managedType.GetBytes(tokenIdHandle)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	esdtRoleKeyPrefix := []byte(core.ElrondProtectedKeyPrefix + core.ESDTRoleIdentifier + core.ESDTKeyIdentifier)
	key := []byte(string(esdtRoleKeyPrefix) + string(tokenID))

	data, usedCache := storage.GetStorage(key)
	storage.UseGasForStorageLoad(storageLoadName, metering.GasSchedule().ElrondAPICost.StorageLoad, usedCache)

	return getESDTRoles(data)
}

// ValidateTokenIdentifier VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) ValidateTokenIdentifier(
	tokenIdHandle int32,
) int32 {
	managedType := context.GetManagedTypesContext()
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetArgument
	metering.UseGasAndAddTracedGas(validateTokenIdentifierName, gasToUse)

	tokenID, err := managedType.GetBytes(tokenIdHandle)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	if ValidateToken(tokenID) {
		return 1
	} else {
		return 0
	}

}

// TransferValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) TransferValue(destOffset int32, valueOffset int32, dataOffset int32, length int32) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	output := host.Output()
	metering.StartGasTracing(transferValueName)

	gasToUse := metering.GasSchedule().ElrondAPICost.TransferValue
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()
	dest, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	valueBytes, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(length))
	metering.UseAndTraceGas(gasToUse)

	data, err := runtime.MemLoad(dataOffset, length)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	if host.IsBuiltinFunctionCall(data) {
		context.WithFault(arwen.ErrTransferValueOnESDTCall, runtime.ElrondAPIErrorShouldFailExecution())
		return 1
	}

	err = output.Transfer(dest, sender, 0, 0, big.NewInt(0).SetBytes(valueBytes), nil, data, vm.DirectCall)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

type indirectContractCallArguments struct {
	dest      []byte
	value     *big.Int
	function  []byte
	args      [][]byte
	actualLen int32
}

func extractIndirectContractCallArgumentsWithValue(
	host arwen.VMHost,
	destOffset int32,
	valueOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) (*indirectContractCallArguments, error) {
	return extractIndirectContractCallArguments(
		host,
		destOffset,
		valueOffset,
		true,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func extractIndirectContractCallArgumentsWithoutValue(
	host arwen.VMHost,
	destOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) (*indirectContractCallArguments, error) {
	return extractIndirectContractCallArguments(
		host,
		destOffset,
		0,
		false,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func extractIndirectContractCallArguments(
	host arwen.VMHost,
	destOffset int32,
	valueOffset int32,
	hasValueOffset bool,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) (*indirectContractCallArguments, error) {
	runtime := host.Runtime()
	metering := host.Metering()

	dest, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if err != nil {
		return nil, err
	}

	var value *big.Int

	if hasValueOffset {
		valueBytes, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
		if err != nil {
			return nil, err
		}
		value = big.NewInt(0).SetBytes(valueBytes)
	}

	function, err := runtime.MemLoad(functionOffset, functionLength)
	if err != nil {
		return nil, err
	}

	args, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
	if err != nil {
		return nil, err
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	return &indirectContractCallArguments{
		dest:      dest,
		value:     value,
		function:  function,
		args:      args,
		actualLen: actualLen,
	}, nil
}

// TransferValueExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) TransferValueExecute(
	destOffset int32,
	valueOffset int32,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	return TransferValueExecuteWithHost(
		host,
		destOffset,
		valueOffset,
		gasLimit,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// TransferValueExecuteWithHost - transferValueExecute with host instead of pointer context
func TransferValueExecuteWithHost(
	host arwen.VMHost,
	destOffset int32,
	valueOffset int32,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(transferValueExecuteName)

	gasToUse := metering.GasSchedule().ElrondAPICost.TransferValue
	metering.UseAndTraceGas(gasToUse)

	callArgs, err := extractIndirectContractCallArgumentsWithValue(
		host, destOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return TransferValueExecuteWithTypedArgs(
		host,
		callArgs.dest,
		callArgs.value,
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferValueExecuteWithTypedArgs - transferValueExecute with args already read from memory
func TransferValueExecuteWithTypedArgs(
	host arwen.VMHost,
	dest []byte,
	value *big.Int,
	gasLimit int64,
	function []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()
	output := host.Output()

	gasToUse := metering.GasSchedule().ElrondAPICost.TransferValue
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()

	var err error
	var contractCallInput *vmcommon.ContractCallInput

	if len(function) > 0 {
		contractCallInput, err = prepareIndirectContractCallInput(
			host,
			sender,
			value,
			gasLimit,
			dest,
			function,
			args,
			gasToUse,
			false,
		)
		if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return 1
		}
	}

	if contractCallInput != nil {
		if host.IsBuiltinFunctionName(contractCallInput.Function) {
			WithFaultAndHost(host, arwen.ErrNilESDTData, runtime.ElrondAPIErrorShouldFailExecution())
			return 1
		}
	}

	if host.AreInSameShard(sender, dest) && contractCallInput != nil && host.Blockchain().IsSmartContract(dest) {
		logEEI.Trace("eGLD pre-transfer execution begin")
		_, err = executeOnDestContextFromAPI(host, contractCallInput)
		if err != nil {
			logEEI.Trace("eGLD pre-transfer execution failed", "error", err)
			WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution())
			return 1
		}

		return 0
	}

	data := ""
	if contractCallInput != nil {
		data = makeCrossShardCallFromInput(contractCallInput.Function, contractCallInput.Arguments)
	}

	metering.UseAndTraceGas(uint64(gasLimit))
	err = output.Transfer(dest, sender, uint64(gasLimit), 0, value, nil, []byte(data), vm.DirectCall)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

func makeCrossShardCallFromInput(function string, arguments [][]byte) string {
	txData := function
	for _, arg := range arguments {
		txData += "@" + hex.EncodeToString(arg)
	}

	return txData
}

// TransferESDTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) TransferESDTExecute(
	destOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	valueOffset int32,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {

	return context.TransferESDTNFTExecute(destOffset, tokenIDOffset, tokenIDLen, valueOffset, 0,
		gasLimit, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
}

// TransferESDTNFTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) TransferESDTNFTExecute(
	destOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	valueOffset int32,
	nonce int64,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(transferESDTNFTExecuteName)
	return TransferESDTNFTExecuteWithHost(
		host,
		destOffset,
		tokenIDOffset,
		tokenIDLen,
		valueOffset,
		nonce,
		gasLimit,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset)
}

// MultiTransferESDTNFTExecute VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) MultiTransferESDTNFTExecute(
	destOffset int32,
	numTokenTransfers int32,
	tokenTransfersArgsLengthOffset int32,
	tokenTransferDataOffset int32,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(multiTransferESDTNFTExecuteName)

	if numTokenTransfers == 0 {
		_ = WithFaultAndHost(host, arwen.ErrFailedTransfer, runtime.ElrondAPIErrorShouldFailExecution())
		return 1
	}

	callArgs, err := extractIndirectContractCallArgumentsWithoutValue(
		host, destOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(callArgs.actualLen))
	metering.UseAndTraceGas(gasToUse)

	transferArgs, actualLen, err := getArgumentsFromMemory(
		host,
		numTokenTransfers*parsers.ArgsPerTransfer,
		tokenTransfersArgsLengthOffset,
		tokenTransferDataOffset,
	)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	transfers := make([]*vmcommon.ESDTTransfer, numTokenTransfers)
	for i := int32(0); i < numTokenTransfers; i++ {
		tokenStartIndex := i * parsers.ArgsPerTransfer
		transfer := &vmcommon.ESDTTransfer{
			ESDTTokenName:  transferArgs[tokenStartIndex],
			ESDTTokenNonce: big.NewInt(0).SetBytes(transferArgs[tokenStartIndex+1]).Uint64(),
			ESDTValue:      big.NewInt(0).SetBytes(transferArgs[tokenStartIndex+2]),
			ESDTTokenType:  uint32(core.Fungible),
		}
		if transfer.ESDTTokenNonce > 0 {
			transfer.ESDTTokenType = uint32(core.NonFungible)
		}
		transfers[i] = transfer
	}

	return TransferESDTNFTExecuteWithTypedArgs(
		host,
		callArgs.dest,
		transfers,
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferESDTNFTExecuteWithHost contains only memory reading of arguments
func TransferESDTNFTExecuteWithHost(
	host arwen.VMHost,
	destOffset int32,
	tokenIDOffset int32,
	tokenIDLen int32,
	valueOffset int32,
	nonce int64,
	gasLimit int64,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	tokenIdentifier, executeErr := runtime.MemLoad(tokenIDOffset, tokenIDLen)
	if WithFaultAndHost(host, executeErr, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	callArgs, err := extractIndirectContractCallArgumentsWithValue(
		host, destOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	gasToUse := math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(callArgs.actualLen))
	metering.UseAndTraceGas(gasToUse)

	transfer := &vmcommon.ESDTTransfer{
		ESDTValue:      callArgs.value,
		ESDTTokenName:  tokenIdentifier,
		ESDTTokenNonce: uint64(nonce),
		ESDTTokenType:  uint32(core.Fungible),
	}
	if nonce > 0 {
		transfer.ESDTTokenType = uint32(core.NonFungible)
	}
	return TransferESDTNFTExecuteWithTypedArgs(
		host,
		callArgs.dest,
		[]*vmcommon.ESDTTransfer{transfer},
		gasLimit,
		callArgs.function,
		callArgs.args,
	)
}

// TransferESDTNFTExecuteWithTypedArgs defines the actual transfer ESDT execute logic
func TransferESDTNFTExecuteWithTypedArgs(
	host arwen.VMHost,
	dest []byte,
	transfers []*vmcommon.ESDTTransfer,
	gasLimit int64,
	function []byte,
	data [][]byte,
) int32 {
	var executeErr error

	runtime := host.Runtime()
	metering := host.Metering()

	output := host.Output()

	gasToUse := metering.GasSchedule().ElrondAPICost.TransferValue * uint64(len(transfers))
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()

	var contractCallInput *vmcommon.ContractCallInput
	if len(function) > 0 {
		contractCallInput, executeErr = prepareIndirectContractCallInput(
			host,
			sender,
			big.NewInt(0),
			gasLimit,
			dest,
			function,
			data,
			gasToUse,
			false,
		)
		if WithFaultAndHost(host, executeErr, runtime.ElrondSyncExecAPIErrorShouldFailExecution()) {
			return 1
		}

		contractCallInput.ESDTTransfers = transfers
	}

	snapshotBeforeTransfer := host.Blockchain().GetSnapshot()

	gasLimitForExec, executeErr := output.TransferESDT(dest, sender, transfers, contractCallInput)
	if WithFaultAndHost(host, executeErr, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	if host.AreInSameShard(sender, dest) && contractCallInput != nil && host.Blockchain().IsSmartContract(dest) {
		contractCallInput.GasProvided = gasLimitForExec
		logEEI.Trace("ESDT post-transfer execution begin")
		_, executeErr := executeOnDestContextFromAPI(host, contractCallInput)
		if executeErr != nil {
			logEEI.Trace("ESDT post-transfer execution failed", "error", executeErr)
			host.Blockchain().RevertToSnapshot(snapshotBeforeTransfer)
			WithFaultAndHost(host, executeErr, runtime.ElrondAPIErrorShouldFailExecution())
			return 1
		}

		return 0
	}

	return 0
}

// CreateAsyncCall VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) CreateAsyncCall(
	destOffset int32,
	valueOffset int32,
	dataOffset int32,
	dataLength int32,
	successOffset int32,
	successLength int32,
	errorOffset int32,
	errorLength int32,
	gas int64,
	extraGasForCallback int64,
) int32 {
	host := context.GetVMHost()
	return CreateAsyncCallWithHost(
		host,
		destOffset,
		valueOffset,
		dataOffset,
		dataLength,
		successOffset,
		successLength,
		errorOffset,
		errorLength,
		gas,
		extraGasForCallback)
}

// CreateAsyncCallWithHost - createAsyncCall with host instead of pointer
func CreateAsyncCallWithHost(host arwen.VMHost,
	destOffset int32,
	valueOffset int32,
	dataOffset int32,
	dataLength int32,
	successOffset int32,
	successLength int32,
	errorOffset int32,
	errorLength int32,
	gas int64,
	extraGasForCallback int64,
) int32 {
	runtime := host.Runtime()

	calledSCAddress, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	data, err := runtime.MemLoad(dataOffset, dataLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	successFunc, err := runtime.MemLoad(successOffset, successLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	errorFunc, err := runtime.MemLoad(errorOffset, errorLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return CreateAsyncCallWithTypedArgs(host,
		calledSCAddress,
		value,
		data,
		successFunc,
		errorFunc,
		gas,
		extraGasForCallback,
		nil)
}

// CreateAsyncCallWithTypedArgs - createAsyncCall with arguments already read from memory
func CreateAsyncCallWithTypedArgs(host arwen.VMHost,
	calledSCAddress []byte,
	value []byte,
	data []byte,
	successFunc []byte,
	errorFunc []byte,
	gas int64,
	extraGasForCallback int64,
	callbackClosure []byte) int32 {

	metering := host.Metering()
	runtime := host.Runtime()
	async := host.Async()

	metering.StartGasTracing(createAsyncCallName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateAsyncCall
	metering.UseAndTraceGas(gasToUse)

	asyncCall := &arwen.AsyncCall{
		Status:          arwen.AsyncCallPending,
		Destination:     calledSCAddress,
		Data:            data,
		ValueBytes:      value,
		GasLimit:        uint64(gas),
		SuccessCallback: string(successFunc),
		ErrorCallback:   string(errorFunc),
		GasLocked:       uint64(extraGasForCallback),
		CallbackClosure: callbackClosure,
	}

	if asyncCall.HasDefinedAnyCallback() {
		gasToUse := metering.GasSchedule().ElrondAPICost.SetAsyncCallback
		metering.UseAndTraceGas(gasToUse)
	}

	err := async.RegisterAsyncCall("", asyncCall)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

// SetAsyncContextCallback VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) SetAsyncContextCallback(
	callback int32,
	callbackLength int32,
	data int32,
	dataLength int32,
	gas int64,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	async := host.Async()
	metering.StartGasTracing(setAsyncContextCallbackName)

	gasToUse := metering.GasSchedule().ElrondAPICost.SetAsyncContextCallback
	metering.UseAndTraceGas(gasToUse)

	callbackNameBytes, err := runtime.MemLoad(callback, callbackLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	dataBytes, err := runtime.MemLoad(data, dataLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	err = async.SetContextCallback(
		string(callbackNameBytes),
		dataBytes,
		uint64(gas))
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

// UpgradeContract VMHooks implementation.
// @autogenerate(VMHooks)
// @autogenerate(VMHooks)
func (context *ElrondApi) UpgradeContract(
	destOffset int32,
	gasLimit int64,
	valueOffset int32,
	codeOffset int32,
	codeMetadataOffset int32,
	length int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(upgradeContractName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateContract
	metering.UseAndTraceGas(gasToUse)

	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	code, err := runtime.MemLoad(codeOffset, length)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	codeMetadata, err := runtime.MemLoad(codeMetadataOffset, arwen.CodeMetadataLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	data, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	calledSCAddress, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	gasSchedule := metering.GasSchedule()
	gasToUse = math.MulUint64(gasSchedule.BaseOperationCost.DataCopyPerByte, uint64(length))
	metering.UseAndTraceGas(gasToUse)

	upgradeContract(host, calledSCAddress, code, codeMetadata, value, data, gasLimit)
}

// UpgradeFromSourceContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) UpgradeFromSourceContract(
	destOffset int32,
	gasLimit int64,
	valueOffset int32,
	sourceContractAddressOffset int32,
	codeMetadataOffset int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(upgradeFromSourceContractName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateContract
	metering.UseAndTraceGas(gasToUse)

	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	sourceContractAddress, err := runtime.MemLoad(sourceContractAddressOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	codeMetadata, err := runtime.MemLoad(codeMetadataOffset, arwen.CodeMetadataLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	data, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	calledSCAddress, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	UpgradeFromSourceContractWithTypedArgs(
		host,
		sourceContractAddress,
		calledSCAddress,
		value,
		data,
		gasLimit,
		codeMetadata,
	)
}

// UpgradeFromSourceContractWithTypedArgs - upgradeFromSourceContract with args already read from memory
func UpgradeFromSourceContractWithTypedArgs(
	host arwen.VMHost,
	sourceContractAddress []byte,
	destContractAddress []byte,
	value []byte,
	data [][]byte,
	gasLimit int64,
	codeMetadata []byte,
) {
	runtime := host.Runtime()
	blockchain := host.Blockchain()

	code, err := blockchain.GetCode(sourceContractAddress)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	upgradeContract(host, destContractAddress, code, codeMetadata, value, data, gasLimit)
}

func upgradeContract(
	host arwen.VMHost,
	destContractAddress []byte,
	code []byte,
	codeMetadata []byte,
	value []byte,
	data [][]byte,
	gasLimit int64,
) {
	runtime := host.Runtime()
	metering := host.Metering()
	gasSchedule := metering.GasSchedule()
	minAsyncCallCost := math.AddUint64(
		math.MulUint64(2, gasSchedule.ElrondAPICost.AsyncCallStep),
		gasSchedule.ElrondAPICost.AsyncCallbackGasLock)
	if uint64(gasLimit) < minAsyncCallCost {
		runtime.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
		return
	}

	// Set up the async call as if it is not known whether the called SC
	// is in the same shard with the caller or not. This will be later resolved
	// by runtime.ExecuteAsyncCall().
	callData := arwen.UpgradeFunctionName + "@" + hex.EncodeToString(code) + "@" + hex.EncodeToString(codeMetadata)
	for _, arg := range data {
		callData += "@" + hex.EncodeToString(arg)
	}

	async := host.Async()
	err := async.RegisterLegacyAsyncCall(
		destContractAddress,
		[]byte(callData),
		value,
	)
	logEEI.Trace("upgradeContract", "error", err)

	if errors.Is(err, arwen.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
		return
	}
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// DeleteContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) DeleteContract(
	destOffset int32,
	gasLimit int64,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(deleteContractName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateContract
	metering.UseAndTraceGas(gasToUse)

	data, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	calledSCAddress, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	deleteContract(
		host,
		calledSCAddress,
		data,
		gasLimit,
	)
}

func deleteContract(
	host arwen.VMHost,
	dest []byte,
	data [][]byte,
	gasLimit int64,
) {
	runtime := host.Runtime()
	metering := host.Metering()
	gasSchedule := metering.GasSchedule()
	minAsyncCallCost := math.AddUint64(
		math.MulUint64(2, gasSchedule.ElrondAPICost.AsyncCallStep),
		gasSchedule.ElrondAPICost.AsyncCallbackGasLock)
	if uint64(gasLimit) < minAsyncCallCost {
		runtime.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
		return
	}

	callData := arwen.DeleteFunctionName
	for _, arg := range data {
		callData += "@" + hex.EncodeToString(arg)
	}

	async := host.Async()
	err := async.RegisterLegacyAsyncCall(
		dest,
		[]byte(callData),
		big.NewInt(0).Bytes(),
	)
	logEEI.Trace("deleteContract", "error", err)

	if errors.Is(err, arwen.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
		return
	}
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// AsyncCall VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) AsyncCall(destOffset int32, valueOffset int32, dataOffset int32, length int32) {
	host := context.GetVMHost()
	runtime := host.Runtime()
	async := host.Async()
	metering := host.Metering()
	metering.StartGasTracing(asyncCallName)

	gasSchedule := metering.GasSchedule()
	gasToUse := gasSchedule.ElrondAPICost.AsyncCallStep
	metering.UseAndTraceGas(gasToUse)

	calledSCAddress, err := runtime.MemLoad(destOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	gasToUse = math.MulUint64(gasSchedule.BaseOperationCost.DataCopyPerByte, uint64(length))
	metering.UseAndTraceGas(gasToUse)

	data, err := runtime.MemLoad(dataOffset, length)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	err = async.RegisterLegacyAsyncCall(calledSCAddress, data, value)
	if errors.Is(err, arwen.ErrNotEnoughGas) {
		runtime.SetRuntimeBreakpointValue(arwen.BreakpointOutOfGas)
		return
	}
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// GetArgumentLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetArgumentLength(id int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetArgument
	metering.UseGasAndAddTracedGas(getArgumentLengthName, gasToUse)

	args := runtime.Arguments()
	if id < 0 || int32(len(args)) <= id {
		context.WithFault(arwen.ErrInvalidArgument, runtime.ElrondAPIErrorShouldFailExecution())
		return -1
	}

	return int32(len(args[id]))
}

// GetArgument VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetArgument(id int32, argOffset int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetArgument
	metering.UseGasAndAddTracedGas(getArgumentName, gasToUse)

	args := runtime.Arguments()
	if id < 0 || int32(len(args)) <= id {
		context.WithFault(arwen.ErrInvalidArgument, runtime.ElrondAPIErrorShouldFailExecution())
		return -1
	}

	err := runtime.MemStore(argOffset, args[id])
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(args[id]))
}

// GetFunction VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetFunction(functionOffset int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetFunction
	metering.UseGasAndAddTracedGas(getFunctionName, gasToUse)

	function := runtime.FunctionName()
	err := runtime.MemStore(functionOffset, []byte(function))
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(function))
}

// GetNumArguments VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetNumArguments() int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetNumArguments
	metering.UseGasAndAddTracedGas(getNumArgumentsName, gasToUse)

	args := runtime.Arguments()
	return int32(len(args))
}

// StorageStore VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) StorageStore(keyOffset int32, keyLength int32, dataOffset int32, dataLength int32) int32 {
	host := context.GetVMHost()
	return StorageStoreWithHost(
		host,
		keyOffset,
		keyLength,
		dataOffset,
		dataLength,
	)
}

// StorageStoreWithHost - storageStore with host instead of pointer context
func StorageStoreWithHost(host arwen.VMHost, keyOffset int32, keyLength int32, dataOffset int32, dataLength int32) int32 {
	runtime := host.Runtime()

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	data, err := runtime.MemLoad(dataOffset, dataLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return StorageStoreWithTypedArgs(host, key, data)
}

// StorageStoreWithTypedArgs - storageStore with args already read from memory
func StorageStoreWithTypedArgs(host arwen.VMHost, key []byte, data []byte) int32 {
	runtime := host.Runtime()
	storage := host.Storage()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.StorageStore
	metering.UseGasAndAddTracedGas(storageStoreName, gasToUse)

	storageStatus, err := storage.SetStorage(key, data)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(storageStatus)
}

// StorageLoadLength VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) StorageLoadLength(keyOffset int32, keyLength int32) int32 {
	runtime := context.GetRuntimeContext()
	storage := context.GetStorageContext()
	metering := context.GetMeteringContext()

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	data, usedCache := storage.GetStorageUnmetered(key)
	storage.UseGasForStorageLoad(storageLoadLengthName, metering.GasSchedule().ElrondAPICost.StorageLoad, usedCache)

	return int32(len(data))
}

// StorageLoadFromAddress VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) StorageLoadFromAddress(addressOffset int32, keyOffset int32, keyLength int32, dataOffset int32) int32 {
	host := context.GetVMHost()
	return StorageLoadFromAddressWithHost(
		host,
		addressOffset,
		keyOffset,
		keyLength,
		dataOffset,
	)
}

// StorageLoadFromAddressWithHost - storageLoadFromAddress with host instead of pointer context
func StorageLoadFromAddressWithHost(host arwen.VMHost, addressOffset int32, keyOffset int32, keyLength int32, dataOffset int32) int32 {
	runtime := host.Runtime()

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	address, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	data := StorageLoadFromAddressWithTypedArgs(host, address, key)

	err = runtime.MemStore(dataOffset, data)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(data))
}

// StorageLoadFromAddressWithTypedArgs - storageLoadFromAddress with args already read from memory
func StorageLoadFromAddressWithTypedArgs(host arwen.VMHost, address []byte, key []byte) []byte {
	storage := host.Storage()
	metering := host.Metering()
	data, usedCache := storage.GetStorageFromAddress(address, key)
	storage.UseGasForStorageLoad(storageLoadFromAddressName, metering.GasSchedule().ElrondAPICost.StorageLoad, usedCache)
	return data
}

// StorageLoad VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) StorageLoad(keyOffset int32, keyLength int32, dataOffset int32) int32 {
	host := context.GetVMHost()
	return StorageLoadWithHost(
		host,
		keyOffset,
		keyLength,
		dataOffset,
	)
}

// StorageLoadWithHost - storageLoad with host instead of pointer context
func StorageLoadWithHost(host arwen.VMHost, keyOffset int32, keyLength int32, dataOffset int32) int32 {
	runtime := host.Runtime()

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	data := StorageLoadWithWithTypedArgs(host, key)

	err = runtime.MemStore(dataOffset, data)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(data))
}

// StorageLoadWithWithTypedArgs - storageLoad with args already read from memory
func StorageLoadWithWithTypedArgs(host arwen.VMHost, key []byte) []byte {
	storage := host.Storage()
	metering := host.Metering()
	data, usedCache := storage.GetStorage(key)
	storage.UseGasForStorageLoad(storageLoadName, metering.GasSchedule().ElrondAPICost.StorageLoad, usedCache)
	return data
}

// SetStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) SetStorageLock(keyOffset int32, keyLength int32, lockTimestamp int64) int32 {
	host := context.GetVMHost()
	return SetStorageLockWithHost(
		host,
		keyOffset,
		keyLength,
		lockTimestamp,
	)
}

// SetStorageLockWithHost - setStorageLock with host instead of pointer context
func SetStorageLockWithHost(host arwen.VMHost, keyOffset int32, keyLength int32, lockTimestamp int64) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.Int64StorageStore
	metering.UseGasAndAddTracedGas(setStorageLockName, gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return SetStorageLockWithTypedArgs(host, key, lockTimestamp)
}

// SetStorageLockWithTypedArgs - setStorageLock with args already read from memory
func SetStorageLockWithTypedArgs(host arwen.VMHost, key []byte, lockTimestamp int64) int32 {
	runtime := host.Runtime()
	storage := host.Storage()
	timeLockKeyPrefix := string(storage.GetVmProtectedPrefix(arwen.TimeLockKeyPrefix))
	timeLockKey := arwen.CustomStorageKey(timeLockKeyPrefix, key)
	bigTimestamp := big.NewInt(0).SetInt64(lockTimestamp)
	storageStatus, err := storage.SetProtectedStorage(timeLockKey, bigTimestamp.Bytes())
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}
	return int32(storageStatus)
}

// GetStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetStorageLock(keyOffset int32, keyLength int32) int64 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	storage := context.GetStorageContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.StorageLoad
	metering.UseGasAndAddTracedGas(getStorageLockName, gasToUse)

	key, err := runtime.MemLoad(keyOffset, keyLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	timeLockKeyPrefix := string(storage.GetVmProtectedPrefix(arwen.TimeLockKeyPrefix))
	timeLockKey := arwen.CustomStorageKey(timeLockKeyPrefix, key)
	data, usedCache := storage.GetStorage(timeLockKey)
	storage.UseGasForStorageLoad(getStorageLockName, metering.GasSchedule().ElrondAPICost.StorageLoad, usedCache)

	timeLock := big.NewInt(0).SetBytes(data).Int64()

	// TODO if timelock <= currentTimeStamp { fail somehow }

	return timeLock
}

// IsStorageLocked VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) IsStorageLocked(keyOffset int32, keyLength int32) int32 {
	timeLock := context.GetStorageLock(keyOffset, keyLength)
	if timeLock < 0 {
		return -1
	}

	currentTimestamp := context.GetBlockTimestamp()
	if timeLock <= currentTimestamp {
		return 0
	}

	return 1
}

// ClearStorageLock VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) ClearStorageLock(keyOffset int32, keyLength int32) int32 {
	return context.SetStorageLock(keyOffset, keyLength, 0)
}

// GetCaller VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCaller(resultOffset int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCaller
	metering.UseGasAndAddTracedGas(getCallerName, gasToUse)

	caller := runtime.GetVMInput().CallerAddr

	err := runtime.MemStore(resultOffset, caller)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}
}

// CheckNoPayment VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) CheckNoPayment() {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(checkNoPaymentName, gasToUse)

	vmInput := runtime.GetVMInput()
	if vmInput.CallValue.Sign() > 0 {
		_ = context.WithFault(arwen.ErrNonPayableFunctionEgld, runtime.ElrondAPIErrorShouldFailExecution())
		return
	}
	if len(vmInput.ESDTTransfers) > 0 {
		_ = context.WithFault(arwen.ErrNonPayableFunctionEsdt, runtime.ElrondAPIErrorShouldFailExecution())
		return
	}
}

// GetCallValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCallValue(resultOffset int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(callValueName, gasToUse)

	value := runtime.GetVMInput().CallValue.Bytes()
	value = arwen.PadBytesLeft(value, arwen.BalanceLen)

	err := runtime.MemStore(resultOffset, value)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(value))
}

// GetESDTValue VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTValue(resultOffset int32) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTValueByIndex(resultOffset, 0)
}

// GetESDTValueByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTValueByIndex(resultOffset int32, index int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getESDTValueByIndexName, gasToUse)

	var value []byte

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	if esdtTransfer != nil && esdtTransfer.ESDTValue.Cmp(arwen.Zero) > 0 {
		value = esdtTransfer.ESDTValue.Bytes()
		value = arwen.PadBytesLeft(value, arwen.BalanceLen)
	}

	err := runtime.MemStore(resultOffset, value)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(value))
}

// GetESDTTokenName VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenName(resultOffset int32) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenNameByIndex(resultOffset, 0)
}

// GetESDTTokenNameByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenNameByIndex(resultOffset int32, index int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getESDTTokenNameByIndexName, gasToUse)

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	var tokenName []byte
	if esdtTransfer != nil {
		tokenName = esdtTransfer.ESDTTokenName
	}

	err := runtime.MemStore(resultOffset, tokenName)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(tokenName))
}

// GetESDTTokenNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenNonce() int64 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenNonceByIndex(0)
}

// GetESDTTokenNonceByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenNonceByIndex(index int32) int64 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getESDTTokenNonceByIndexName, gasToUse)

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	nonce := uint64(0)
	if esdtTransfer != nil {
		nonce = esdtTransfer.ESDTTokenNonce
	}
	return int64(nonce)
}

// GetCurrentESDTNFTNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCurrentESDTNFTNonce(addressOffset int32, tokenIDOffset int32, tokenIDLen int32) int64 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()
	storage := context.GetStorageContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.StorageLoad
	metering.UseGasAndAddTracedGas(getCurrentESDTNFTNonceName, gasToUse)

	destination, err := runtime.MemLoad(addressOffset, arwen.AddressLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 0
	}

	tokenID, err := runtime.MemLoad(tokenIDOffset, tokenIDLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 0
	}

	key := []byte(core.ElrondProtectedKeyPrefix + core.ESDTNFTLatestNonceIdentifier + string(tokenID))
	data, _ := storage.GetStorageFromAddress(destination, key)

	nonce := big.NewInt(0).SetBytes(data).Uint64()
	return int64(nonce)
}

// GetESDTTokenType VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenType() int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetESDTTokenTypeByIndex(0)
}

// GetESDTTokenTypeByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetESDTTokenTypeByIndex(index int32) int32 {
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getESDTTokenTypeByIndexName, gasToUse)

	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)
	if esdtTransfer != nil {
		return int32(esdtTransfer.ESDTTokenType)
	}
	return 0
}

// GetNumESDTTransfers VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetNumESDTTransfers() int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getNumESDTTransfersName, gasToUse)

	return int32(len(runtime.GetVMInput().ESDTTransfers))
}

// GetCallValueTokenName VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCallValueTokenName(callValueOffset int32, tokenNameOffset int32) int32 {
	isFail := failIfMoreThanOneESDTTransfer(context)
	if isFail {
		return -1
	}
	return context.GetCallValueTokenNameByIndex(callValueOffset, tokenNameOffset, 0)
}

// GetCallValueTokenNameByIndex VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCallValueTokenNameByIndex(callValueOffset int32, tokenNameOffset int32, index int32) int32 {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCallValue
	metering.UseGasAndAddTracedGas(getCallValueTokenNameByIndexName, gasToUse)

	callValue := runtime.GetVMInput().CallValue.Bytes()
	tokenName := make([]byte, 0)
	esdtTransfer := getESDTTransferFromInputFailIfWrongIndex(context.GetVMHost(), index)

	if esdtTransfer != nil {
		tokenName = make([]byte, len(esdtTransfer.ESDTTokenName))
		copy(tokenName, esdtTransfer.ESDTTokenName)
		callValue = esdtTransfer.ESDTValue.Bytes()
	}
	callValue = arwen.PadBytesLeft(callValue, arwen.BalanceLen)

	err := runtime.MemStore(tokenNameOffset, tokenName)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	err = runtime.MemStore(callValueOffset, callValue)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return int32(len(tokenName))
}

// WriteLog VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) WriteLog(dataPointer int32, dataLength int32, topicPtr int32, numTopics int32) {
	// note: deprecated
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.Log
	gas := math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(numTopics*arwen.HashLen+dataLength))
	gasToUse = math.AddUint64(gasToUse, gas)
	metering.UseGasAndAddTracedGas(writeLogName, gasToUse)

	if numTopics < 0 || dataLength < 0 {
		err := arwen.ErrNegativeLength
		context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
		return
	}

	log, err := runtime.MemLoad(dataPointer, dataLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	topics := make([][]byte, numTopics)
	for i := int32(0); i < numTopics; i++ {
		topics[i], err = runtime.MemLoad(topicPtr+i*arwen.HashLen, arwen.HashLen)
		if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
			return
		}
	}

	output.WriteLog(runtime.GetContextAddress(), topics, log)
}

// WriteEventLog VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) WriteEventLog(
	numTopics int32,
	topicLengthsOffset int32,
	topicOffset int32,
	dataOffset int32,
	dataLength int32,
) {

	host := context.GetVMHost()
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	topics, topicDataTotalLen, err := getArgumentsFromMemory(
		host,
		numTopics,
		topicLengthsOffset,
		topicOffset,
	)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	data, err := runtime.MemLoad(dataOffset, dataLength)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	gasToUse := metering.GasSchedule().ElrondAPICost.Log
	gasForData := math.MulUint64(
		metering.GasSchedule().BaseOperationCost.DataCopyPerByte,
		uint64(topicDataTotalLen+dataLength))
	gasToUse = math.AddUint64(gasToUse, gasForData)
	metering.UseGasAndAddTracedGas(writeEventLogName, gasToUse)

	output.WriteLog(runtime.GetContextAddress(), topics, data)
}

// GetBlockTimestamp VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockTimestamp() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockTimeStamp
	metering.UseGasAndAddTracedGas(getBlockTimestampName, gasToUse)

	return int64(blockchain.CurrentTimeStamp())
}

// GetBlockNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockNonce() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockNonce
	metering.UseGasAndAddTracedGas(getBlockNonceName, gasToUse)

	return int64(blockchain.CurrentNonce())
}

// GetBlockRound VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockRound() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockRound
	metering.UseGasAndAddTracedGas(getBlockRoundName, gasToUse)

	return int64(blockchain.CurrentRound())
}

// GetBlockEpoch VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockEpoch() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockEpoch
	metering.UseGasAndAddTracedGas(getBlockEpochName, gasToUse)

	return int64(blockchain.CurrentEpoch())
}

// GetBlockRandomSeed VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetBlockRandomSeed(pointer int32) {
	runtime := context.GetRuntimeContext()
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockRandomSeed
	metering.UseGasAndAddTracedGas(getBlockRandomSeedName, gasToUse)

	randomSeed := blockchain.CurrentRandomSeed()
	err := runtime.MemStore(pointer, randomSeed)
	context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

// GetStateRootHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetStateRootHash(pointer int32) {
	runtime := context.GetRuntimeContext()
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetStateRootHash
	metering.UseGasAndAddTracedGas(getStateRootHashName, gasToUse)

	stateRootHash := blockchain.GetStateRootHash()
	err := runtime.MemStore(pointer, stateRootHash)
	context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

// GetPrevBlockTimestamp VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevBlockTimestamp() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockTimeStamp
	metering.UseGasAndAddTracedGas(getPrevBlockTimestampName, gasToUse)

	return int64(blockchain.LastTimeStamp())
}

// GetPrevBlockNonce VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevBlockNonce() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockNonce
	metering.UseGasAndAddTracedGas(getPrevBlockNonceName, gasToUse)

	return int64(blockchain.LastNonce())
}

// GetPrevBlockRound VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevBlockRound() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockRound
	metering.UseGasAndAddTracedGas(getPrevBlockRoundName, gasToUse)

	return int64(blockchain.LastRound())
}

// GetPrevBlockEpoch VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevBlockEpoch() int64 {
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockEpoch
	metering.UseGasAndAddTracedGas(getPrevBlockEpochName, gasToUse)

	return int64(blockchain.LastEpoch())
}

// GetPrevBlockRandomSeed VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevBlockRandomSeed(pointer int32) {
	runtime := context.GetRuntimeContext()
	blockchain := context.GetBlockchainContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetBlockRandomSeed
	metering.UseGasAndAddTracedGas(getPrevBlockRandomSeedName, gasToUse)

	randomSeed := blockchain.LastRandomSeed()
	err := runtime.MemStore(pointer, randomSeed)
	context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

// Finish VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) Finish(pointer int32, length int32) {
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()
	metering.StartGasTracing(returnDataName)

	gasToUse := metering.GasSchedule().ElrondAPICost.Finish
	gas := math.MulUint64(metering.GasSchedule().BaseOperationCost.PersistPerByte, uint64(length))
	gasToUse = math.AddUint64(gasToUse, gas)
	err := metering.UseGasBounded(gasToUse)

	if err != nil {
		_ = context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
		return
	}

	data, err := runtime.MemLoad(pointer, length)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return
	}

	output.Finish(data)
}

// ExecuteOnSameContext VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) ExecuteOnSameContext(
	gasLimit int64,
	addressOffset int32,
	valueOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeOnSameContextName)

	return ExecuteOnSameContextWithHost(
		host,
		gasLimit,
		addressOffset,
		valueOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteOnSameContextWithHost - executeOnSameContext with host instead of pointer context
func ExecuteOnSameContextWithHost(
	host arwen.VMHost,
	gasLimit int64,
	addressOffset int32,
	valueOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()

	callArgs, err := extractIndirectContractCallArgumentsWithValue(
		host, addressOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return ExecuteOnSameContextWithTypedArgs(
		host,
		gasLimit,
		callArgs.value,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteOnSameContextWithTypedArgs - executeOnSameContext with args already read from memory
func ExecuteOnSameContextWithTypedArgs(
	host arwen.VMHost,
	gasLimit int64,
	value *big.Int,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.ExecuteOnSameContext
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()
	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		value,
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	if host.IsBuiltinFunctionName(contractCallInput.Function) {
		WithFaultAndHost(host, arwen.ErrInvalidBuiltInFunctionCall, runtime.ElrondAPIErrorShouldFailExecution())
		return 1
	}

	err = host.ExecuteOnSameContext(contractCallInput)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return 0
}

// ExecuteOnDestContext VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) ExecuteOnDestContext(
	gasLimit int64,
	addressOffset int32,
	valueOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeOnDestContextName)

	return ExecuteOnDestContextWithHost(
		host,
		gasLimit,
		addressOffset,
		valueOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteOnDestContextWithHost - executeOnDestContext with host instead of pointer context
func ExecuteOnDestContextWithHost(
	host arwen.VMHost,
	gasLimit int64,
	addressOffset int32,
	valueOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()

	callArgs, err := extractIndirectContractCallArgumentsWithValue(
		host, addressOffset, valueOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return ExecuteOnDestContextWithTypedArgs(
		host,
		gasLimit,
		callArgs.value,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteOnDestContextWithTypedArgs - executeOnDestContext with args already read from memory
func ExecuteOnDestContextWithTypedArgs(
	host arwen.VMHost,
	gasLimit int64,
	value *big.Int,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.ExecuteOnDestContext
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()
	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		value,
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	_, err = executeOnDestContextFromAPI(host, contractCallInput)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

// ExecuteReadOnly VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) ExecuteReadOnly(
	gasLimit int64,
	addressOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	metering := host.Metering()
	metering.StartGasTracing(executeReadOnlyName)

	return ExecuteReadOnlyWithHost(
		host,
		gasLimit,
		addressOffset,
		functionOffset,
		functionLength,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

// ExecuteReadOnlyWithHost - executeReadOnly with host instead of pointer context
func ExecuteReadOnlyWithHost(
	host arwen.VMHost,
	gasLimit int64,
	addressOffset int32,
	functionOffset int32,
	functionLength int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()

	callArgs, err := extractIndirectContractCallArgumentsWithoutValue(
		host, addressOffset, functionOffset, functionLength, numArguments, argumentsLengthOffset, dataOffset)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return ExecuteReadOnlyWithTypedArguments(
		host,
		gasLimit,
		callArgs.function,
		callArgs.dest,
		callArgs.args,
	)
}

// ExecuteReadOnlyWithTypedArguments - executeReadOnly with args already read from memory
func ExecuteReadOnlyWithTypedArguments(
	host arwen.VMHost,
	gasLimit int64,
	function []byte,
	dest []byte,
	args [][]byte,
) int32 {
	runtime := host.Runtime()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.ExecuteReadOnly
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()
	contractCallInput, err := prepareIndirectContractCallInput(
		host,
		sender,
		big.NewInt(0),
		gasLimit,
		dest,
		function,
		args,
		gasToUse,
		true,
	)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	if host.IsBuiltinFunctionName(contractCallInput.Function) {
		WithFaultAndHost(host, arwen.ErrInvalidBuiltInFunctionCall, runtime.ElrondAPIErrorShouldFailExecution())
		return 1
	}

	wasReadOnly := runtime.ReadOnly()
	runtime.SetReadOnly(true)
	_, err = executeOnDestContextFromAPI(host, contractCallInput)
	runtime.SetReadOnly(wasReadOnly)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return -1
	}

	return 0
}

// CreateContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) CreateContract(
	gasLimit int64,
	valueOffset int32,
	codeOffset int32,
	codeMetadataOffset int32,
	length int32,
	resultOffset int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	return createContractWithHost(
		host,
		gasLimit,
		valueOffset,
		codeOffset,
		codeMetadataOffset,
		length,
		resultOffset,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)
}

func createContractWithHost(
	host arwen.VMHost,
	gasLimit int64,
	valueOffset int32,
	codeOffset int32,
	codeMetadataOffset int32,
	length int32,
	resultOffset int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	runtime := host.Runtime()

	metering := host.Metering()
	metering.StartGasTracing(createContractName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateContract
	metering.UseAndTraceGas(gasToUse)

	sender := runtime.GetContextAddress()
	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	code, err := runtime.MemLoad(codeOffset, length)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	codeMetadata, err := runtime.MemLoad(codeMetadataOffset, arwen.CodeMetadataLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	data, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	valueAsInt := big.NewInt(0).SetBytes(value)
	newAddress, err := createContract(sender, data, valueAsInt, metering, gasLimit, code, codeMetadata, host, runtime)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	err = runtime.MemStore(resultOffset, newAddress)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

// DeployFromSourceContract VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) DeployFromSourceContract(
	gasLimit int64,
	valueOffset int32,
	sourceContractAddressOffset int32,
	codeMetadataOffset int32,
	resultAddressOffset int32,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) int32 {
	host := context.GetVMHost()
	runtime := host.Runtime()
	metering := host.Metering()
	metering.StartGasTracing(deployFromSourceContractName)

	gasToUse := metering.GasSchedule().ElrondAPICost.CreateContract
	metering.UseAndTraceGas(gasToUse)

	value, err := runtime.MemLoad(valueOffset, arwen.BalanceLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	sourceContractAddress, err := runtime.MemLoad(sourceContractAddressOffset, arwen.AddressLen)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	codeMetadata, err := runtime.MemLoad(codeMetadataOffset, arwen.CodeMetadataLen)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	data, actualLen, err := getArgumentsFromMemory(
		host,
		numArguments,
		argumentsLengthOffset,
		dataOffset,
	)

	gasToUse = math.MulUint64(metering.GasSchedule().BaseOperationCost.DataCopyPerByte, uint64(actualLen))
	metering.UseAndTraceGas(gasToUse)

	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	newAddress, err := DeployFromSourceContractWithTypedArgs(
		host,
		sourceContractAddress,
		codeMetadata,
		big.NewInt(0).SetBytes(value),
		data,
		gasLimit,
	)

	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	err = runtime.MemStore(resultAddressOffset, newAddress)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 1
	}

	return 0
}

// DeployFromSourceContractWithTypedArgs - deployFromSourceContract with args already read from memory
func DeployFromSourceContractWithTypedArgs(
	host arwen.VMHost,
	sourceContractAddress []byte,
	codeMetadata []byte,
	value *big.Int,
	data [][]byte,
	gasLimit int64,
) ([]byte, error) {
	runtime := host.Runtime()
	metering := host.Metering()
	sender := runtime.GetContextAddress()

	blockchain := host.Blockchain()
	code, err := blockchain.GetCode(sourceContractAddress)
	if WithFaultAndHost(host, err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return nil, err
	}

	return createContract(sender, data, value, metering, gasLimit, code, codeMetadata, host, runtime)
}

func createContract(
	sender []byte,
	data [][]byte,
	value *big.Int,
	metering arwen.MeteringContext,
	gasLimit int64,
	code []byte,
	codeMetadata []byte,
	host arwen.VMHost,
	_ arwen.RuntimeContext,
) ([]byte, error) {
	contractCreate := &vmcommon.ContractCreateInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   data,
			CallValue:   value,
			GasPrice:    0,
			GasProvided: metering.BoundGasLimit(gasLimit),
		},
		ContractCode:         code,
		ContractCodeMetadata: codeMetadata,
	}

	return host.CreateNewContract(contractCreate)
}

// GetNumReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetNumReturnData() int32 {
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetNumReturnData
	metering.UseGasAndAddTracedGas(getNumReturnDataName, gasToUse)

	returnData := output.ReturnData()
	return int32(len(returnData))
}

// GetReturnDataSize VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetReturnDataSize(resultID int32) int32 {
	runtime := context.GetRuntimeContext()
	output := context.GetOutputContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetReturnDataSize
	metering.UseGasAndAddTracedGas(getReturnDataSizeName, gasToUse)

	returnData := output.ReturnData()
	if resultID >= int32(len(returnData)) || resultID < 0 {
		context.WithFault(arwen.ErrInvalidArgument, runtime.ElrondAPIErrorShouldFailExecution())
		return 0
	}

	return int32(len(returnData[resultID]))
}

// GetReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetReturnData(resultID int32, dataOffset int32) int32 {
	host := context.GetVMHost()

	result := GetReturnDataWithHostAndTypedArgs(host, resultID)
	if result == nil {
		return 0
	}

	runtime := context.GetRuntimeContext()
	err := runtime.MemStore(dataOffset, result)
	if context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution()) {
		return 0
	}

	return int32(len(result))
}

func GetReturnDataWithHostAndTypedArgs(host arwen.VMHost, resultID int32) []byte {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetReturnData
	metering.UseGasAndAddTracedGas(getReturnDataName, gasToUse)

	returnData := output.ReturnData()
	if resultID >= int32(len(returnData)) || resultID < 0 {
		WithFaultAndHost(host, arwen.ErrInvalidArgument, host.Runtime().ElrondAPIErrorShouldFailExecution())
		return nil
	}

	return returnData[resultID]
}

// CleanReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) CleanReturnData() {
	host := context.GetVMHost()
	CleanReturnDataWithHost(host)
}

// CleanReturnDataWithHost - exposed version of v1_5_deleteFromReturnData for tests
func CleanReturnDataWithHost(host arwen.VMHost) {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.CleanReturnData
	metering.UseGasAndAddTracedGas(cleanReturnDataName, gasToUse)

	output.ClearReturnData()
}

// DeleteFromReturnData VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) DeleteFromReturnData(resultID int32) {
	host := context.GetVMHost()
	DeleteFromReturnDataWithHost(host, resultID)
}

// DeleteFromReturnDataWithHost - exposed version of v1_5_deleteFromReturnData for tests
func DeleteFromReturnDataWithHost(host arwen.VMHost, resultID int32) {
	output := host.Output()
	metering := host.Metering()

	gasToUse := metering.GasSchedule().ElrondAPICost.DeleteFromReturnData
	metering.UseGasAndAddTracedGas(deleteFromReturnDataName, gasToUse)

	returnData := output.ReturnData()
	if resultID < int32(len(returnData)) {
		output.RemoveReturnData(uint32(resultID))
	}
}

// GetOriginalTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetOriginalTxHash(dataOffset int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetOriginalTxHash
	metering.UseGasAndAddTracedGas(getOriginalTxHashName, gasToUse)

	err := runtime.MemStore(dataOffset, runtime.GetOriginalTxHash())
	_ = context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

// GetCurrentTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetCurrentTxHash(dataOffset int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetCurrentTxHash
	metering.UseGasAndAddTracedGas(getCurrentTxHashName, gasToUse)

	err := runtime.MemStore(dataOffset, runtime.GetCurrentTxHash())
	_ = context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

// GetPrevTxHash VMHooks implementation.
// @autogenerate(VMHooks)
func (context *ElrondApi) GetPrevTxHash(dataOffset int32) {
	runtime := context.GetRuntimeContext()
	metering := context.GetMeteringContext()

	gasToUse := metering.GasSchedule().ElrondAPICost.GetPrevTxHash
	metering.UseGasAndAddTracedGas(getPrevTxHashName, gasToUse)

	err := runtime.MemStore(dataOffset, runtime.GetPrevTxHash())
	_ = context.WithFault(err, runtime.ElrondAPIErrorShouldFailExecution())
}

func prepareIndirectContractCallInput(
	host arwen.VMHost,
	sender []byte,
	value *big.Int,
	gasLimit int64,
	destination []byte,
	function []byte,
	data [][]byte,
	_ uint64,
	syncExecutionRequired bool,
) (*vmcommon.ContractCallInput, error) {
	runtime := host.Runtime()
	metering := host.Metering()

	if syncExecutionRequired && !host.AreInSameShard(runtime.GetContextAddress(), destination) {
		return nil, arwen.ErrSyncExecutionNotInSameShard
	}

	contractCallInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   data,
			CallValue:   value,
			GasPrice:    0,
			GasProvided: metering.BoundGasLimit(gasLimit),
			CallType:    vm.DirectCall,
		},
		RecipientAddr: destination,
		Function:      string(function),
	}

	return contractCallInput, nil
}

func getArgumentsFromMemory(
	host arwen.VMHost,
	numArguments int32,
	argumentsLengthOffset int32,
	dataOffset int32,
) ([][]byte, int32, error) {
	runtime := host.Runtime()

	if numArguments < 0 {
		return nil, 0, fmt.Errorf("negative numArguments (%d)", numArguments)
	}

	argumentsLengthData, err := runtime.MemLoad(argumentsLengthOffset, numArguments*4)
	if err != nil {
		return nil, 0, err
	}

	argumentLengths := createInt32Array(argumentsLengthData, numArguments)
	data, err := runtime.MemLoadMultiple(dataOffset, argumentLengths)
	if err != nil {
		return nil, 0, err
	}

	totalArgumentBytes := int32(0)
	for _, length := range argumentLengths {
		totalArgumentBytes += length
	}

	return data, totalArgumentBytes, nil
}

func createInt32Array(rawData []byte, numIntegers int32) []int32 {
	integers := make([]int32, numIntegers)
	index := 0
	for cursor := 0; cursor < len(rawData); cursor += 4 {
		rawInt := rawData[cursor : cursor+4]
		actualInt := binary.LittleEndian.Uint32(rawInt)
		integers[index] = int32(actualInt)
		index++
	}
	return integers
}

func executeOnDestContextFromAPI(host arwen.VMHost, input *vmcommon.ContractCallInput) (vmOutput *vmcommon.VMOutput, err error) {
	host.Async().SetAsyncArgumentsForCall(input)
	vmOutput, isChildComplete, err := host.ExecuteOnDestContext(input)
	if err != nil {
		return nil, err
	}
	err = host.Async().CompleteChildConditional(isChildComplete, nil, 0)
	if err != nil {
		return nil, err
	}
	return vmOutput, err
}
