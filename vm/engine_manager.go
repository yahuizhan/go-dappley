package vm

import (
	"strings"

	"github.com/dappley/go-dappley/core/client"
	"github.com/dappley/go-dappley/core"
)

const scheduleFuncName = "dapp_schedule"

type V8EngineManager struct {
	address client.Address
}

func NewV8EngineManager(address client.Address) *V8EngineManager {
	return &V8EngineManager{address}
}

func (em *V8EngineManager) CreateEngine() core.ScEngine {
	engine := NewV8Engine()
	engine.ImportNodeAddress(em.address)
	return engine
}

func (em *V8EngineManager) RunScheduledEvents(contractUtxos []*core.UTXO,
	scStorage *core.ScState,
	blkHeight uint64,
	seed int64) {

	for _, utxo := range contractUtxos {
		if !strings.Contains(utxo.Contract, scheduleFuncName) {
			continue
		}
		addr := utxo.PubKeyHash.GenerateAddress()

		engine := em.CreateEngine()
		// TODO confirm whether we need to set limit
		if err := engine.SetExecutionLimits(DefaultLimitsOfGas, DefaultLimitsOfTotalMemorySize); err != nil {
			continue
		}
		engine.ImportSourceCode(utxo.Contract)
		engine.ImportLocalStorage(scStorage)
		engine.ImportContractAddr(addr)
		engine.ImportSourceTXID(utxo.Txid)
		engine.ImportCurrBlockHeight(blkHeight)
		engine.ImportSeed(seed)
		engine.Execute(scheduleFuncName, "")
		engine.DestroyEngine()
	}
}
