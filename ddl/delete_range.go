// Copyright 2017 PingCAP, Inc.
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

package ddl

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/pingcap/errors"
	"github.com/pingcap/failpoint"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	sess "github.com/pingcap/tidb/ddl/internal/session"
	"github.com/pingcap/tidb/ddl/util"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/terror"
	"github.com/pingcap/tidb/sessionctx"
	"github.com/pingcap/tidb/tablecodec"
	"github.com/pingcap/tidb/util/logutil"
	"github.com/pingcap/tidb/util/sqlexec"
	topsqlstate "github.com/pingcap/tidb/util/topsql/state"
	"go.uber.org/zap"
)

const (
	insertDeleteRangeSQLPrefix = `INSERT IGNORE INTO mysql.gc_delete_range VALUES `
	insertDeleteRangeSQLValue  = `(%?, %?, %?, %?, %?)`
	insertDeleteRangeSQL       = insertDeleteRangeSQLPrefix + insertDeleteRangeSQLValue

	delBatchSize = 65536
	delBackLog   = 128
)

var (
	// batchInsertDeleteRangeSize is the maximum size for each batch insert statement in the delete-range.
	batchInsertDeleteRangeSize = 256
)

type delRangeManager interface {
	// addDelRangeJob add a DDL job into gc_delete_range table.
	addDelRangeJob(ctx context.Context, job *model.Job) error
	// removeFromGCDeleteRange removes the deleting table job from gc_delete_range table by jobID and tableID.
	// It's use for recover the table that was mistakenly deleted.
	removeFromGCDeleteRange(ctx context.Context, jobID int64) error
	start()
	clear()
}

type delRange struct {
	store      kv.Storage
	sessPool   *sess.Pool
	emulatorCh chan struct{}
	keys       []kv.Key
	quitCh     chan struct{}

	wait         sync.WaitGroup // wait is only used when storeSupport is false.
	storeSupport bool
}

// newDelRangeManager returns a delRangeManager.
func newDelRangeManager(store kv.Storage, sessPool *sess.Pool) delRangeManager {
	dr := &delRange{
		store:        store,
		sessPool:     sessPool,
		storeSupport: store.SupportDeleteRange(),
		quitCh:       make(chan struct{}),
	}
	if !dr.storeSupport {
		dr.emulatorCh = make(chan struct{}, delBackLog)
		dr.keys = make([]kv.Key, 0, delBatchSize)
	}
	return dr
}

// addDelRangeJob implements delRangeManager interface.
func (dr *delRange) addDelRangeJob(ctx context.Context, job *model.Job) error {
	sctx, err := dr.sessPool.Get()
	if err != nil {
		return errors.Trace(err)
	}
	defer dr.sessPool.Put(sctx)

	if job.MultiSchemaInfo != nil {
		err = insertJobIntoDeleteRangeTableMultiSchema(ctx, sctx, job)
	} else {
		err = insertJobIntoDeleteRangeTable(ctx, sctx, job, &elementIDAlloc{})
	}
	if err != nil {
		logutil.BgLogger().Error("add job into delete-range table failed", zap.String("category", "ddl"), zap.Int64("jobID", job.ID), zap.String("jobType", job.Type.String()), zap.Error(err))
		return errors.Trace(err)
	}
	if !dr.storeSupport {
		dr.emulatorCh <- struct{}{}
	}
	logutil.BgLogger().Info("add job into delete-range table", zap.String("category", "ddl"), zap.Int64("jobID", job.ID), zap.String("jobType", job.Type.String()))
	return nil
}

func insertJobIntoDeleteRangeTableMultiSchema(ctx context.Context, sctx sessionctx.Context, job *model.Job) error {
	var ea elementIDAlloc
	for _, sub := range job.MultiSchemaInfo.SubJobs {
		proxyJob := sub.ToProxyJob(job)
		if jobNeedGC(&proxyJob) {
			err := insertJobIntoDeleteRangeTable(ctx, sctx, &proxyJob, &ea)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

// removeFromGCDeleteRange implements delRangeManager interface.
func (dr *delRange) removeFromGCDeleteRange(ctx context.Context, jobID int64) error {
	sctx, err := dr.sessPool.Get()
	if err != nil {
		return errors.Trace(err)
	}
	defer dr.sessPool.Put(sctx)
	err = util.RemoveMultiFromGCDeleteRange(ctx, sctx, jobID)
	return errors.Trace(err)
}

// start implements delRangeManager interface.
func (dr *delRange) start() {
	if !dr.storeSupport {
		dr.wait.Add(1)
		go dr.startEmulator()
	}
}

// clear implements delRangeManager interface.
func (dr *delRange) clear() {
	logutil.BgLogger().Info("closing delRange", zap.String("category", "ddl"))
	close(dr.quitCh)
	dr.wait.Wait()
}

// startEmulator is only used for those storage engines which don't support
// delete-range. The emulator fetches records from gc_delete_range table and
// deletes all keys in each DelRangeTask.
func (dr *delRange) startEmulator() {
	defer dr.wait.Done()
	logutil.BgLogger().Info("start delRange emulator", zap.String("category", "ddl"))
	for {
		select {
		case <-dr.emulatorCh:
		case <-dr.quitCh:
			return
		}
		if util.IsEmulatorGCEnable() {
			err := dr.doDelRangeWork()
			terror.Log(errors.Trace(err))
		}
	}
}

func (dr *delRange) doDelRangeWork() error {
	sctx, err := dr.sessPool.Get()
	if err != nil {
		logutil.BgLogger().Error("delRange emulator get session failed", zap.String("category", "ddl"), zap.Error(err))
		return errors.Trace(err)
	}
	defer dr.sessPool.Put(sctx)

	ctx := kv.WithInternalSourceType(context.Background(), kv.InternalTxnDDL)
	ranges, err := util.LoadDeleteRanges(ctx, sctx, math.MaxInt64)
	if err != nil {
		logutil.BgLogger().Error("delRange emulator load tasks failed", zap.String("category", "ddl"), zap.Error(err))
		return errors.Trace(err)
	}

	for _, r := range ranges {
		if err := dr.doTask(sctx, r); err != nil {
			logutil.BgLogger().Error("delRange emulator do task failed", zap.String("category", "ddl"), zap.Error(err))
			return errors.Trace(err)
		}
	}
	return nil
}

func (dr *delRange) doTask(sctx sessionctx.Context, r util.DelRangeTask) error {
	var oldStartKey, newStartKey kv.Key
	oldStartKey = r.StartKey
	for {
		finish := true
		dr.keys = dr.keys[:0]
		ctx := kv.WithInternalSourceType(context.Background(), kv.InternalTxnDDL)
		err := kv.RunInNewTxn(ctx, dr.store, false, func(ctx context.Context, txn kv.Transaction) error {
			if topsqlstate.TopSQLEnabled() {
				// Only when TiDB run without PD(use unistore as storage for test) will run into here, so just set a mock internal resource tagger.
				txn.SetOption(kv.ResourceGroupTagger, util.GetInternalResourceGroupTaggerForTopSQL())
			}
			iter, err := txn.Iter(oldStartKey, r.EndKey)
			if err != nil {
				return errors.Trace(err)
			}
			defer iter.Close()

			txn.SetDiskFullOpt(kvrpcpb.DiskFullOpt_AllowedOnAlmostFull)
			for i := 0; i < delBatchSize; i++ {
				if !iter.Valid() {
					break
				}
				finish = false
				dr.keys = append(dr.keys, iter.Key().Clone())
				newStartKey = iter.Key().Next()

				if err := iter.Next(); err != nil {
					return errors.Trace(err)
				}
			}

			for _, key := range dr.keys {
				err := txn.Delete(key)
				if err != nil && !kv.ErrNotExist.Equal(err) {
					return errors.Trace(err)
				}
			}
			return nil
		})
		if err != nil {
			return errors.Trace(err)
		}
		if finish {
			if err := util.CompleteDeleteRange(sctx, r, true); err != nil {
				logutil.BgLogger().Error("delRange emulator complete task failed", zap.String("category", "ddl"), zap.Error(err))
				return errors.Trace(err)
			}
			startKey, endKey := r.Range()
			logutil.BgLogger().Info("delRange emulator complete task", zap.String("category", "ddl"),
				zap.Int64("jobID", r.JobID),
				zap.Int64("elementID", r.ElementID),
				zap.Stringer("startKey", startKey),
				zap.Stringer("endKey", endKey))
			break
		}
		if err := util.UpdateDeleteRange(sctx, r, newStartKey, oldStartKey); err != nil {
			logutil.BgLogger().Error("delRange emulator update task failed", zap.String("category", "ddl"), zap.Error(err))
		}
		oldStartKey = newStartKey
	}
	return nil
}

// insertJobIntoDeleteRangeTable parses the job into delete-range arguments,
// and inserts a new record into gc_delete_range table. The primary key is
// (job ID, element ID), so we ignore key conflict error.
func insertJobIntoDeleteRangeTable(ctx context.Context, sctx sessionctx.Context, job *model.Job, ea *elementIDAlloc) error {
	now, err := getNowTSO(sctx)
	if err != nil {
		return errors.Trace(err)
	}

	ctx = kv.WithInternalSourceType(ctx, getDDLRequestSource(job.Type))
	s := sctx.(sqlexec.SQLExecutor)
	switch job.Type {
	case model.ActionDropSchema:
		var tableIDs []int64
		if err := job.DecodeArgs(&tableIDs); err != nil {
			return errors.Trace(err)
		}
		for i := 0; i < len(tableIDs); i += batchInsertDeleteRangeSize {
			batchEnd := len(tableIDs)
			if batchEnd > i+batchInsertDeleteRangeSize {
				batchEnd = i + batchInsertDeleteRangeSize
			}
			if err := doBatchInsert(ctx, s, job.ID, tableIDs[i:batchEnd], now, ea); err != nil {
				return errors.Trace(err)
			}
		}
	case model.ActionDropTable, model.ActionTruncateTable:
		tableID := job.TableID
		// The startKey here is for compatibility with previous versions, old version did not endKey so don't have to deal with.
		var startKey kv.Key
		var physicalTableIDs []int64
		var ruleIDs []string
		if err := job.DecodeArgs(&startKey, &physicalTableIDs, &ruleIDs); err != nil {
			return errors.Trace(err)
		}
		if len(physicalTableIDs) > 0 {
			for _, pid := range physicalTableIDs {
				startKey = tablecodec.EncodeTablePrefix(pid)
				endKey := tablecodec.EncodeTablePrefix(pid + 1)
				elemID := ea.allocForPhysicalID(pid)
				if err := doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("partition ID is %d", pid)); err != nil {
					return errors.Trace(err)
				}
			}
			// logical table may contain global index regions, so delete the logical table range.
			startKey = tablecodec.EncodeTablePrefix(tableID)
			endKey := tablecodec.EncodeTablePrefix(tableID + 1)
			elemID := ea.allocForPhysicalID(tableID)
			return doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("table ID is %d", tableID))
		}
		startKey = tablecodec.EncodeTablePrefix(tableID)
		endKey := tablecodec.EncodeTablePrefix(tableID + 1)
		elemID := ea.allocForPhysicalID(tableID)
		return doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("table ID is %d", tableID))
	case model.ActionDropTablePartition, model.ActionTruncateTablePartition,
		model.ActionReorganizePartition, model.ActionRemovePartitioning,
		model.ActionAlterTablePartitioning:
		var physicalTableIDs []int64
		// partInfo is not used, but is set in ReorgPartition.
		// Better to have an additional argument in job.DecodeArgs since it is ignored,
		// instead of having one to few, which will remove the data from the job arguments...
		var partInfo model.PartitionInfo
		if err := job.DecodeArgs(&physicalTableIDs, &partInfo); err != nil {
			return errors.Trace(err)
		}
		for _, physicalTableID := range physicalTableIDs {
			startKey := tablecodec.EncodeTablePrefix(physicalTableID)
			endKey := tablecodec.EncodeTablePrefix(physicalTableID + 1)
			elemID := ea.allocForPhysicalID(physicalTableID)
			if err := doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("partition table ID is %d", physicalTableID)); err != nil {
				return errors.Trace(err)
			}
		}
	// ActionAddIndex, ActionAddPrimaryKey needs do it, because it needs to be rolled back when it's canceled.
	case model.ActionAddIndex, model.ActionAddPrimaryKey:
		allIndexIDs := make([]int64, 1)
		ifExists := make([]bool, 1)
		var partitionIDs []int64
		if err := job.DecodeArgs(&allIndexIDs[0], &ifExists[0], &partitionIDs); err != nil {
			if err = job.DecodeArgs(&allIndexIDs, &ifExists, &partitionIDs); err != nil {
				return errors.Trace(err)
			}
		}
		// Determine the physicalIDs to be added.
		physicalIDs := []int64{job.TableID}
		if len(partitionIDs) > 0 {
			physicalIDs = partitionIDs
		}
		for _, indexID := range allIndexIDs {
			// Determine the index IDs to be added.
			tempIdxID := tablecodec.TempIndexPrefix | indexID
			var indexIDs []int64
			if job.State == model.JobStateRollbackDone {
				indexIDs = []int64{indexID, tempIdxID}
			} else {
				indexIDs = []int64{tempIdxID}
			}
			for _, pid := range physicalIDs {
				for _, iid := range indexIDs {
					startKey := tablecodec.EncodeTableIndexPrefix(pid, iid)
					endKey := tablecodec.EncodeTableIndexPrefix(pid, iid+1)
					elemID := ea.allocForIndexID(pid, iid)
					if err := doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("physical ID is %d", pid)); err != nil {
						return errors.Trace(err)
					}
				}
			}
		}
	case model.ActionDropIndex, model.ActionDropPrimaryKey:
		tableID := job.TableID
		var indexName interface{}
		var partitionIDs []int64
		ifExists := make([]bool, 1)
		allIndexIDs := make([]int64, 1)
		if err := job.DecodeArgs(&indexName, &ifExists[0], &allIndexIDs[0], &partitionIDs); err != nil {
			if err = job.DecodeArgs(&indexName, &ifExists, &allIndexIDs, &partitionIDs); err != nil {
				return errors.Trace(err)
			}
		}
		for _, indexID := range allIndexIDs {
			// partitionIDs len is 0 if the dropped index is a global index, even if it is a partitioned table.
			if len(partitionIDs) == 0 {
				startKey := tablecodec.EncodeTableIndexPrefix(tableID, indexID)
				endKey := tablecodec.EncodeTableIndexPrefix(tableID, indexID+1)
				elemID := ea.allocForIndexID(tableID, indexID)
				return doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("index ID is %d", indexID))
			}
			failpoint.Inject("checkDropGlobalIndex", func(val failpoint.Value) {
				if val.(bool) {
					panic("drop global index must not delete partition index range")
				}
			})
			for _, pid := range partitionIDs {
				startKey := tablecodec.EncodeTableIndexPrefix(pid, indexID)
				endKey := tablecodec.EncodeTableIndexPrefix(pid, indexID+1)
				elemID := ea.allocForIndexID(pid, indexID)
				if err := doInsert(ctx, s, job.ID, elemID, startKey, endKey, now, fmt.Sprintf("partition table ID is %d", pid)); err != nil {
					return errors.Trace(err)
				}
			}
		}
	case model.ActionDropColumn:
		var colName model.CIStr
		var ifExists bool
		var indexIDs []int64
		var partitionIDs []int64
		if err := job.DecodeArgs(&colName, &ifExists, &indexIDs, &partitionIDs); err != nil {
			return errors.Trace(err)
		}
		if len(indexIDs) > 0 {
			if len(partitionIDs) == 0 {
				return doBatchDeleteIndiceRange(ctx, s, job.ID, job.TableID, indexIDs, now, ea)
			}
			for _, pid := range partitionIDs {
				if err := doBatchDeleteIndiceRange(ctx, s, job.ID, pid, indexIDs, now, ea); err != nil {
					return errors.Trace(err)
				}
			}
		}
	case model.ActionModifyColumn:
		var indexIDs []int64
		var partitionIDs []int64
		if err := job.DecodeArgs(&indexIDs, &partitionIDs); err != nil {
			return errors.Trace(err)
		}
		if len(indexIDs) == 0 {
			return nil
		}
		if len(partitionIDs) == 0 {
			return doBatchDeleteIndiceRange(ctx, s, job.ID, job.TableID, indexIDs, now, ea)
		}
		for _, pid := range partitionIDs {
			if err := doBatchDeleteIndiceRange(ctx, s, job.ID, pid, indexIDs, now, ea); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func doBatchDeleteIndiceRange(ctx context.Context, s sqlexec.SQLExecutor, jobID, tableID int64, indexIDs []int64, ts uint64, ea *elementIDAlloc) error {
	logutil.BgLogger().Info("batch insert into delete-range indices", zap.String("category", "ddl"), zap.Int64("jobID", jobID), zap.Int64("tableID", tableID), zap.Int64s("indexIDs", indexIDs))
	paramsList := make([]interface{}, 0, len(indexIDs)*5)
	var buf strings.Builder
	buf.WriteString(insertDeleteRangeSQLPrefix)
	for i, indexID := range indexIDs {
		startKey := tablecodec.EncodeTableIndexPrefix(tableID, indexID)
		endKey := tablecodec.EncodeTableIndexPrefix(tableID, indexID+1)
		startKeyEncoded := hex.EncodeToString(startKey)
		endKeyEncoded := hex.EncodeToString(endKey)
		buf.WriteString(insertDeleteRangeSQLValue)
		if i != len(indexIDs)-1 {
			buf.WriteString(",")
		}
		elemID := ea.allocForIndexID(tableID, indexID)
		paramsList = append(paramsList, jobID, elemID, startKeyEncoded, endKeyEncoded, ts)
	}
	_, err := s.ExecuteInternal(ctx, buf.String(), paramsList...)
	return errors.Trace(err)
}

func doInsert(ctx context.Context, s sqlexec.SQLExecutor, jobID, elementID int64, startKey, endKey kv.Key, ts uint64, comment string) error {
	logutil.BgLogger().Info("insert into delete-range table", zap.String("category", "ddl"), zap.Int64("jobID", jobID), zap.Int64("elementID", elementID), zap.String("comment", comment))
	startKeyEncoded := hex.EncodeToString(startKey)
	endKeyEncoded := hex.EncodeToString(endKey)
	// set session disk full opt
	// TODO ddl txn func including an session pool txn, there may be a problem?
	s.SetDiskFullOpt(kvrpcpb.DiskFullOpt_AllowedOnAlmostFull)
	_, err := s.ExecuteInternal(ctx, insertDeleteRangeSQL, jobID, elementID, startKeyEncoded, endKeyEncoded, ts)
	// clear session disk full opt
	s.ClearDiskFullOpt()
	return errors.Trace(err)
}

func doBatchInsert(ctx context.Context, s sqlexec.SQLExecutor, jobID int64, tableIDs []int64, ts uint64, ea *elementIDAlloc) error {
	logutil.BgLogger().Info("batch insert into delete-range table", zap.String("category", "ddl"), zap.Int64("jobID", jobID), zap.Int64s("tableIDs", tableIDs))
	var buf strings.Builder
	buf.WriteString(insertDeleteRangeSQLPrefix)
	paramsList := make([]interface{}, 0, len(tableIDs)*5)
	for i, tableID := range tableIDs {
		startKey := tablecodec.EncodeTablePrefix(tableID)
		endKey := tablecodec.EncodeTablePrefix(tableID + 1)
		startKeyEncoded := hex.EncodeToString(startKey)
		endKeyEncoded := hex.EncodeToString(endKey)
		buf.WriteString(insertDeleteRangeSQLValue)
		if i != len(tableIDs)-1 {
			buf.WriteString(",")
		}
		elemID := ea.allocForPhysicalID(tableID)
		paramsList = append(paramsList, jobID, elemID, startKeyEncoded, endKeyEncoded, ts)
	}
	// set session disk full opt
	s.SetDiskFullOpt(kvrpcpb.DiskFullOpt_AllowedOnAlmostFull)
	_, err := s.ExecuteInternal(ctx, buf.String(), paramsList...)
	// clear session disk full opt
	s.ClearDiskFullOpt()
	return errors.Trace(err)
}

// getNowTS gets the current timestamp, in TSO.
func getNowTSO(ctx sessionctx.Context) (uint64, error) {
	currVer, err := ctx.GetStore().CurrentVersion(kv.GlobalTxnScope)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return currVer.Ver, nil
}
