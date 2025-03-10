// Copyright 2023 PingCAP, Inc.
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

package storage

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/sessionctx/stmtctx"
	"github.com/pingcap/tidb/statistics"
	"github.com/pingcap/tidb/types"
	compressutil "github.com/pingcap/tidb/util/compress"
	"github.com/pingcap/tipb/go-tipb"
)

// JSONTable is used for dumping statistics.
type JSONTable struct {
	Columns           map[string]*jsonColumn `json:"columns"`
	Indices           map[string]*jsonColumn `json:"indices"`
	Partitions        map[string]*JSONTable  `json:"partitions"`
	DatabaseName      string                 `json:"database_name"`
	TableName         string                 `json:"table_name"`
	ExtStats          []*JSONExtendedStats   `json:"ext_stats"`
	Count             int64                  `json:"count"`
	ModifyCount       int64                  `json:"modify_count"`
	Version           uint64                 `json:"version"`
	IsHistoricalStats bool                   `json:"is_historical_stats"`
}

// JSONExtendedStats is used for dumping extended statistics.
type JSONExtendedStats struct {
	StatsName  string  `json:"stats_name"`
	StringVals string  `json:"string_vals"`
	ColIDs     []int64 `json:"cols"`
	ScalarVals float64 `json:"scalar_vals"`
	Tp         uint8   `json:"type"`
}

func dumpJSONExtendedStats(statsColl *statistics.ExtendedStatsColl) []*JSONExtendedStats {
	if statsColl == nil || len(statsColl.Stats) == 0 {
		return nil
	}
	stats := make([]*JSONExtendedStats, 0, len(statsColl.Stats))
	for name, item := range statsColl.Stats {
		js := &JSONExtendedStats{
			StatsName:  name,
			ColIDs:     item.ColIDs,
			Tp:         item.Tp,
			ScalarVals: item.ScalarVals,
			StringVals: item.StringVals,
		}
		stats = append(stats, js)
	}
	return stats
}

func extendedStatsFromJSON(statsColl []*JSONExtendedStats) *statistics.ExtendedStatsColl {
	if len(statsColl) == 0 {
		return nil
	}
	stats := statistics.NewExtendedStatsColl()
	for _, js := range statsColl {
		item := &statistics.ExtendedStatsItem{
			ColIDs:     js.ColIDs,
			Tp:         js.Tp,
			ScalarVals: js.ScalarVals,
			StringVals: js.StringVals,
		}
		stats.Stats[js.StatsName] = item
	}
	return stats
}

type jsonColumn struct {
	Histogram *tipb.Histogram `json:"histogram"`
	CMSketch  *tipb.CMSketch  `json:"cm_sketch"`
	FMSketch  *tipb.FMSketch  `json:"fm_sketch"`
	// StatsVer is a pointer here since the old version json file would not contain version information.
	StatsVer          *int64  `json:"stats_ver"`
	NullCount         int64   `json:"null_count"`
	TotColSize        int64   `json:"tot_col_size"`
	LastUpdateVersion uint64  `json:"last_update_version"`
	Correlation       float64 `json:"correlation"`
}

func dumpJSONCol(hist *statistics.Histogram, cmsketch *statistics.CMSketch, topn *statistics.TopN, fmsketch *statistics.FMSketch, statsVer *int64) *jsonColumn {
	jsonCol := &jsonColumn{
		Histogram:         statistics.HistogramToProto(hist),
		NullCount:         hist.NullCount,
		TotColSize:        hist.TotColSize,
		LastUpdateVersion: hist.LastUpdateVersion,
		Correlation:       hist.Correlation,
		StatsVer:          statsVer,
	}
	if cmsketch != nil || topn != nil {
		jsonCol.CMSketch = statistics.CMSketchToProto(cmsketch, topn)
	}
	if fmsketch != nil {
		jsonCol.FMSketch = statistics.FMSketchToProto(fmsketch)
	}
	return jsonCol
}

// GenJSONTableFromStats generate jsonTable from tableInfo and stats
func GenJSONTableFromStats(dbName string, tableInfo *model.TableInfo, tbl *statistics.Table) (*JSONTable, error) {
	jsonTbl := &JSONTable{
		DatabaseName: dbName,
		TableName:    tableInfo.Name.L,
		Columns:      make(map[string]*jsonColumn, len(tbl.Columns)),
		Indices:      make(map[string]*jsonColumn, len(tbl.Indices)),
		Count:        tbl.RealtimeCount,
		ModifyCount:  tbl.ModifyCount,
		Version:      tbl.Version,
	}
	for _, col := range tbl.Columns {
		sc := &stmtctx.StatementContext{TimeZone: time.UTC}
		hist, err := col.ConvertTo(sc, types.NewFieldType(mysql.TypeBlob))
		if err != nil {
			return nil, errors.Trace(err)
		}
		jsonTbl.Columns[col.Info.Name.L] = dumpJSONCol(hist, col.CMSketch, col.TopN, col.FMSketch, &col.StatsVer)
	}

	for _, idx := range tbl.Indices {
		jsonTbl.Indices[idx.Info.Name.L] = dumpJSONCol(&idx.Histogram, idx.CMSketch, idx.TopN, nil, &idx.StatsVer)
	}
	jsonTbl.ExtStats = dumpJSONExtendedStats(tbl.ExtendedStats)
	return jsonTbl, nil
}

// TableStatsFromJSON loads statistic from JSONTable and return the Table of statistic.
func TableStatsFromJSON(tableInfo *model.TableInfo, physicalID int64, jsonTbl *JSONTable) (*statistics.Table, error) {
	newHistColl := statistics.HistColl{
		PhysicalID:     physicalID,
		HavePhysicalID: true,
		RealtimeCount:  jsonTbl.Count,
		ModifyCount:    jsonTbl.ModifyCount,
		Columns:        make(map[int64]*statistics.Column, len(jsonTbl.Columns)),
		Indices:        make(map[int64]*statistics.Index, len(jsonTbl.Indices)),
	}
	tbl := &statistics.Table{
		HistColl: newHistColl,
	}
	for id, jsonIdx := range jsonTbl.Indices {
		for _, idxInfo := range tableInfo.Indices {
			if idxInfo.Name.L != id {
				continue
			}
			hist := statistics.HistogramFromProto(jsonIdx.Histogram)
			hist.ID, hist.NullCount, hist.LastUpdateVersion, hist.Correlation = idxInfo.ID, jsonIdx.NullCount, jsonIdx.LastUpdateVersion, jsonIdx.Correlation
			cm, topN := statistics.CMSketchAndTopNFromProto(jsonIdx.CMSketch)
			statsVer := int64(statistics.Version0)
			if jsonIdx.StatsVer != nil {
				statsVer = *jsonIdx.StatsVer
			} else if jsonIdx.Histogram.Ndv > 0 || jsonIdx.NullCount > 0 {
				// If the statistics are collected without setting stats version(which happens in v4.0 and earlier versions),
				// we set it to 1.
				statsVer = int64(statistics.Version1)
			}
			idx := &statistics.Index{
				Histogram:         *hist,
				CMSketch:          cm,
				TopN:              topN,
				Info:              idxInfo,
				StatsVer:          statsVer,
				PhysicalID:        physicalID,
				StatsLoadedStatus: statistics.NewStatsFullLoadStatus(),
			}
			tbl.Indices[idx.ID] = idx
		}
	}

	for id, jsonCol := range jsonTbl.Columns {
		for _, colInfo := range tableInfo.Columns {
			if colInfo.Name.L != id {
				continue
			}
			hist := statistics.HistogramFromProto(jsonCol.Histogram)
			sc := &stmtctx.StatementContext{TimeZone: time.UTC}
			tmpFT := colInfo.FieldType
			// For new collation data, when storing the bounds of the histogram, we store the collate key instead of the
			// original value.
			// But there's additional conversion logic for new collation data, and the collate key might be longer than
			// the FieldType.flen.
			// If we use the original FieldType here, there might be errors like "Invalid utf8mb4 character string"
			// or "Data too long".
			// So we change it to TypeBlob to bypass those logics here.
			if colInfo.FieldType.EvalType() == types.ETString && colInfo.FieldType.GetType() != mysql.TypeEnum && colInfo.FieldType.GetType() != mysql.TypeSet {
				tmpFT = *types.NewFieldType(mysql.TypeBlob)
			}
			hist, err := hist.ConvertTo(sc, &tmpFT)
			if err != nil {
				return nil, errors.Trace(err)
			}
			cm, topN := statistics.CMSketchAndTopNFromProto(jsonCol.CMSketch)
			fms := statistics.FMSketchFromProto(jsonCol.FMSketch)
			hist.ID, hist.NullCount, hist.LastUpdateVersion, hist.TotColSize, hist.Correlation = colInfo.ID, jsonCol.NullCount, jsonCol.LastUpdateVersion, jsonCol.TotColSize, jsonCol.Correlation
			statsVer := int64(statistics.Version0)
			if jsonCol.StatsVer != nil {
				statsVer = *jsonCol.StatsVer
			} else if jsonCol.Histogram.Ndv > 0 || jsonCol.NullCount > 0 {
				// If the statistics are collected without setting stats version(which happens in v4.0 and earlier versions),
				// we set it to 1.
				statsVer = int64(statistics.Version1)
			}
			col := &statistics.Column{
				PhysicalID:        physicalID,
				Histogram:         *hist,
				CMSketch:          cm,
				TopN:              topN,
				FMSketch:          fms,
				Info:              colInfo,
				IsHandle:          tableInfo.PKIsHandle && mysql.HasPriKeyFlag(colInfo.GetFlag()),
				StatsVer:          statsVer,
				StatsLoadedStatus: statistics.NewStatsFullLoadStatus(),
			}
			tbl.Columns[col.ID] = col
		}
	}
	tbl.ExtendedStats = extendedStatsFromJSON(jsonTbl.ExtStats)
	return tbl, nil
}

// JSONTableToBlocks convert JSONTable to json, then compresses it to blocks by gzip.
func JSONTableToBlocks(jsTable *JSONTable, blockSize int) ([][]byte, error) {
	data, err := json.Marshal(jsTable)
	if err != nil {
		return nil, errors.Trace(err)
	}
	var gzippedData bytes.Buffer
	gzipWriter := compressutil.GzipWriterPool.Get().(*gzip.Writer)
	defer compressutil.GzipWriterPool.Put(gzipWriter)
	gzipWriter.Reset(&gzippedData)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, errors.Trace(err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, errors.Trace(err)
	}
	blocksNum := gzippedData.Len() / blockSize
	if gzippedData.Len()%blockSize != 0 {
		blocksNum = blocksNum + 1
	}
	blocks := make([][]byte, blocksNum)
	for i := 0; i < blocksNum-1; i++ {
		blocks[i] = gzippedData.Bytes()[blockSize*i : blockSize*(i+1)]
	}
	blocks[blocksNum-1] = gzippedData.Bytes()[blockSize*(blocksNum-1):]
	return blocks, nil
}

// BlocksToJSONTable convert gzip-compressed blocks to JSONTable
func BlocksToJSONTable(blocks [][]byte) (*JSONTable, error) {
	if len(blocks) == 0 {
		return nil, errors.New("Block empty error")
	}
	data := blocks[0]
	for i := 1; i < len(blocks); i++ {
		data = append(data, blocks[i]...)
	}
	gzippedData := bytes.NewReader(data)
	gzipReader := compressutil.GzipReaderPool.Get().(*gzip.Reader)
	if err := gzipReader.Reset(gzippedData); err != nil {
		compressutil.GzipReaderPool.Put(gzipReader)
		return nil, err
	}
	defer func() {
		compressutil.GzipReaderPool.Put(gzipReader)
	}()
	if err := gzipReader.Close(); err != nil {
		return nil, err
	}
	jsonStr, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, errors.Trace(err)
	}
	jsonTbl := JSONTable{}
	err = json.Unmarshal(jsonStr, &jsonTbl)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &jsonTbl, nil
}
