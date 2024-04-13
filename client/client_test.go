package client

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/concrete-eth/archetype/arch"
	"github.com/concrete-eth/archetype/kvstore"
	"github.com/concrete-eth/archetype/testutils"
	"github.com/concrete-eth/archetype/utils"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/log"
)

func newTestClient(t *testing.T) (*Client, lib.KeyValueStore, chan arch.ActionBatch, chan []arch.Action) {
	var (
		specs                    = testutils.NewTestArchSpecs(t)
		core                     = &testutils.Core{}
		kv                       = kvstore.NewMemoryKeyValueStore()
		actionBatchInChan        = make(chan arch.ActionBatch)
		actionOutChan            = make(chan []arch.Action)
		blockTime                = 10 * time.Millisecond
		blockNumber       uint64 = 0
	)
	client, err := New(specs, core, kv, actionBatchInChan, actionOutChan, blockTime, blockNumber)
	if err != nil {
		t.Fatal(err)
	}
	return client, kv, actionBatchInChan, actionOutChan
}

func TestSimulate(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	client.Simulate(func(_core arch.Core) {
		core := _core.(*testutils.Core)
		core.SetCounter(1)
		if c := core.GetCounter(); c != 1 {
			t.Errorf("expected %v, got %v", 1, c)
		}
	})
	if c := client.Core.(*testutils.Core).GetCounter(); c != 0 {
		t.Errorf("expected %v, got %v", 0, c)
	}
}

func TestSendActions(t *testing.T) {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlError, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	client, _, _, actionOutChan := newTestClient(t)
	actionsIn := []arch.Action{&arch.CanonicalTickAction{}, &testutils.ActionData_Add{}}
	go client.SendActions(actionsIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionOutChan:
		if !reflect.DeepEqual(actionsIn, actionsOut) {
			t.Fatal("unexpected actions")
		}
	}
}

func TestSendAction(t *testing.T) {
	client, _, _, actionOutChan := newTestClient(t)
	actionIn := &testutils.ActionData_Add{}
	go client.SendAction(actionIn)
	select {
	case <-time.After(10 * time.Millisecond):
		t.Fatal("timeout")
	case actionsOut := <-actionOutChan:
		if len(actionsOut) != 1 {
			t.Fatal("unexpected actions")
		}
		if !reflect.DeepEqual(actionIn, actionsOut[0]) {
			t.Fatal("unexpected action")
		}
	}
}

var testData = []struct {
	batch                arch.ActionBatch
	expTickActionInBatch bool
	expCounterValue      int16
}{
	{
		arch.ActionBatch{
			BlockNumber: 0,
			Actions: []arch.Action{
				&arch.CanonicalTickAction{},
				&testutils.ActionData_Add{Summand: 1},
			},
		},
		true,
		1,
	},
	{
		arch.ActionBatch{
			BlockNumber: 1,
			Actions: []arch.Action{
				&arch.CanonicalTickAction{},
			},
		},
		true,
		4,
	},
	{
		arch.ActionBatch{
			BlockNumber: 2,
			Actions: []arch.Action{
				&testutils.ActionData_Add{Summand: 1},
			},
		},
		false,
		5,
	},
	{
		arch.ActionBatch{
			BlockNumber: 3,
			Actions:     []arch.Action{},
		},
		false,
		5,
	},
}

func init() {
	for i, data := range testData {
		if data.batch.BlockNumber != uint64(i) {
			panic("unexpected block number")
		}
	}
}

func TestSync(t *testing.T) {
	client, _, _, _ := newTestClient(t)
	actionBatchInChan := make(chan arch.ActionBatch, 1)
	client.actionBatchInChan = actionBatchInChan

	didReceiveNewBatch, didTick, err := client.Sync()
	if err != nil {
		t.Fatal(err)
	}
	if didReceiveNewBatch {
		t.Fatal("unexpected new batch")
	}
	if didTick {
		t.Fatal("expected no tick action")
	}
	if client.Core.BlockNumber() != 0 {
		t.Fatal("unexpected block number")
	}
	if client.Core.(*testutils.Core).GetCounter() != 0 {
		t.Fatal("unexpected value")
	}

	for _, actionBatch := range testData {
		actionBatchInChan <- actionBatch.batch
		_, didTick, err = client.Sync()
		if err != nil {
			t.Fatal(err)
		}
		if didTick != actionBatch.expTickActionInBatch {
			t.Errorf("expected %v, got %v", actionBatch.expTickActionInBatch, didTick)
		}
		if client.Core.BlockNumber() != actionBatch.batch.BlockNumber+1 {
			t.Errorf("expected %v, got %v", actionBatch.batch.BlockNumber+1, client.Core.BlockNumber())
		}
		if c := client.Core.(*testutils.Core).GetCounter(); c != actionBatch.expCounterValue {
			t.Errorf("expected %v, got %v", actionBatch.expCounterValue, c)
		}
	}
}

func TestSyncUntil(t *testing.T) {
	client, _, actionBatchInChan, _ := newTestClient(t)

	err := client.SyncUntil(0)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for _, actionBatch := range testData {
			actionBatchInChan <- actionBatch.batch
		}
	}()

	for _, blockToSyncTo := range []uint64{2, 4} {
		err := client.SyncUntil(blockToSyncTo)
		if err != nil {
			t.Fatal(err)
		}
		if client.Core.BlockNumber() != blockToSyncTo {
			t.Fatal("unexpected block number")
		}
		if c := client.Core.(*testutils.Core).GetCounter(); c != testData[blockToSyncTo-1].expCounterValue {
			t.Errorf("expected %v, got %v", testData[blockToSyncTo-1].expCounterValue, c)
		}
	}
}

func TestInterpolatedSync(t *testing.T) {
	client, _, actionBatchInChan, _ := newTestClient(t)

	didReceiveNewBatch, didTick, err := client.InterpolatedSync()
	if err != nil {
		t.Fatal(err)
	}
	if didReceiveNewBatch {
		t.Fatal("unexpected new batch")
	}
	if !didTick {
		t.Fatal("expected tick")
	}

	go func() {
		ticker := time.NewTicker(client.blockTime)
		defer ticker.Stop()
		for _, data := range testData {
			<-ticker.C
			actionBatchInChan <- data.batch
		}
	}()

	var (
		ticksPerBlock = client.Core.TicksPerBlock()
		tickPeriod    = client.blockTime / time.Duration(ticksPerBlock)
	)

	ticker := time.NewTicker(client.blockTime / 4)

	for range ticker.C {
		didReceiveNewBatch, _, err := client.InterpolatedSync()
		if err != nil {
			t.Fatal(err)
		}
		var expectedCoreVal int16
		if client.Core.BlockNumber() == 0 {
			expectedCoreVal = 0
		} else {
			expectedCoreVal = testData[client.Core.BlockNumber()-1].expCounterValue
		}
		if !didReceiveNewBatch {
			// Adjust expectedCoreVal for interpolated ticks
			targetTicks := uint(time.Since(client.lastNewBatchTime)/tickPeriod) + 1
			targetTicks = utils.Min(targetTicks, ticksPerBlock)
			expectedCoreVal *= int16(2 * targetTicks)
		}
		if c := client.Core.(*testutils.Core).GetCounter(); c != expectedCoreVal {
			t.Errorf("expected %v, got %v", expectedCoreVal, c)
		}
		if int(client.Core.BlockNumber()) >= len(testData) {
			break
		}
	}

	if c := client.Core.(*testutils.Core).GetCounter(); c != testData[len(testData)-1].expCounterValue {
		t.Errorf("expected %v, got %v", testData[len(testData)-1].expCounterValue, c)
	}
}
