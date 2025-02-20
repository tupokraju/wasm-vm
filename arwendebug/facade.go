package arwendebug

import (
	"encoding/json"
	"fmt"
	"io"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("arwendebug")

// DebugFacade is the debug facade
type DebugFacade struct {
}

// NewDebugFacade creates a new debug facade
func NewDebugFacade() *DebugFacade {
	return &DebugFacade{}
}

// DeploySmartContract deploys a smart contract
func (f *DebugFacade) DeploySmartContract(request DeployRequest) (*DeployResponse, error) {
	log.Debug("DeploySmartContract()")

	err := request.digest()
	if err != nil {
		return nil, err
	}

	database := f.loadDatabase(request.DatabasePath)
	world, err := database.loadWorld(request.World)
	if err != nil {
		return nil, err
	}
	defer func() {
		vmAsClose := world.vm.(io.Closer)
		_ = vmAsClose.Close()
	}()

	response := world.deploySmartContract(request)

	err = database.storeWorld(world)
	if err != nil {
		return nil, err
	}

	err = database.storeOutcome(request.Outcome, response)
	if err != nil {
		return nil, err
	}

	dumpOutcome(&response)
	return response, err
}

func (f *DebugFacade) loadDatabase(rootPath string) *database {
	database := newDatabase(rootPath)
	return database
}

// UpgradeSmartContract upgrades a smart contract
func (f *DebugFacade) UpgradeSmartContract(request UpgradeRequest) (*UpgradeResponse, error) {
	log.Debug("UpgradeSmartContract()")

	err := request.digest()
	if err != nil {
		return nil, err
	}

	database := f.loadDatabase(request.DatabasePath)
	world, err := database.loadWorld(request.World)
	if err != nil {
		return nil, err
	}
	defer func() {
		vmAsClose := world.vm.(io.Closer)
		_ = vmAsClose.Close()
	}()

	response := world.upgradeSmartContract(request)

	err = database.storeWorld(world)
	if err != nil {
		return nil, err
	}

	err = database.storeOutcome(request.Outcome, response)
	if err != nil {
		return nil, err
	}

	dumpOutcome(&response)

	return response, err
}

// RunSmartContract executes a smart contract function
func (f *DebugFacade) RunSmartContract(request RunRequest) (*RunResponse, error) {
	log.Debug("RunSmartContract()")

	err := request.digest()
	if err != nil {
		return nil, err
	}

	database := f.loadDatabase(request.DatabasePath)
	world, err := database.loadWorld(request.World)
	if err != nil {
		return nil, err
	}
	defer func() {
		vmAsClose := world.vm.(io.Closer)
		_ = vmAsClose.Close()
	}()

	response := world.runSmartContract(request)

	err = database.storeWorld(world)
	if err != nil {
		return nil, err
	}

	err = database.storeOutcome(request.Outcome, response)
	if err != nil {
		return nil, err
	}

	dumpOutcome(&response)
	return response, err
}

// QuerySmartContract queries a pure function of the smart contract
func (f *DebugFacade) QuerySmartContract(request QueryRequest) (*QueryResponse, error) {
	log.Debug("QuerySmartContracts()")

	err := request.digest()
	if err != nil {
		return nil, err
	}

	database := f.loadDatabase(request.DatabasePath)
	world, err := database.loadWorld(request.World)
	if err != nil {
		return nil, err
	}
	defer func() {
		vmAsClose := world.vm.(io.Closer)
		_ = vmAsClose.Close()
	}()

	response := world.querySmartContract(request)

	err = database.storeOutcome(request.Outcome, response)
	if err != nil {
		return nil, err
	}

	dumpOutcome(&response)
	return response, err
}

// CreateAccount creates a test account
func (f *DebugFacade) CreateAccount(request CreateAccountRequest) (*CreateAccountResponse, error) {
	log.Debug("CreateAccount()")

	err := request.digest()
	if err != nil {
		return nil, err
	}

	database := f.loadDatabase(request.DatabasePath)
	world, err := database.loadWorld(request.World)
	if err != nil {
		return nil, err
	}
	defer func() {
		vmAsClose := world.vm.(io.Closer)
		_ = vmAsClose.Close()
	}()

	response := world.createAccount(request)

	err = database.storeWorld(world)
	if err != nil {
		return nil, err
	}

	err = database.storeOutcome(request.Outcome, response)
	if err != nil {
		return nil, err
	}

	dumpOutcome(&response)

	return response, err
}

func dumpOutcome(outcome interface{}) {
	data, err := json.MarshalIndent(outcome, "", "\t")
	if err != nil {
		fmt.Println("{}")
	}

	fmt.Println(string(data))
}
