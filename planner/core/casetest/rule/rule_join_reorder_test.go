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

package rule

import (
	"testing"

	"github.com/pingcap/failpoint"
	"github.com/pingcap/tidb/domain"
	"github.com/pingcap/tidb/parser/model"
	"github.com/pingcap/tidb/testkit"
	"github.com/pingcap/tidb/testkit/testdata"
	"github.com/stretchr/testify/require"
)

func runJoinReorderTestData(t *testing.T, tk *testkit.TestKit, name string) {
	var input []string
	var output []struct {
		SQL     string
		Plan    []string
		Warning []string
	}
	joinReorderSuiteData := GetJoinReorderSuiteData()
	joinReorderSuiteData.LoadTestCasesByName(name, t, &input, &output)
	require.Equal(t, len(input), len(output))
	for i := range input {
		testdata.OnRecord(func() {
			output[i].SQL = input[i]
			output[i].Plan = testdata.ConvertRowsToStrings(tk.MustQuery("explain format = 'brief' " + input[i]).Rows())
			output[i].Warning = testdata.ConvertRowsToStrings(tk.MustQuery("show warnings").Rows())
		})
		tk.MustQuery("explain format = 'brief' " + input[i]).Check(testkit.Rows(output[i].Plan...))
		tk.MustQuery("show warnings").Check(testkit.Rows(output[i].Warning...))
	}
}

func TestStraightJoinHint(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	runJoinReorderTestData(t, tk, "TestStraightJoinHint")
}

func TestNoHashJoinHint(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	runJoinReorderTestData(t, tk, "TestNoHashJoinHint")
}

// test the global/session variable tidb_opt_enable_hash_join being set to no
func TestOptEnableHashJoin(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_opt_enable_hash_join=off")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	runJoinReorderTestData(t, tk, "TestOptEnableHashJoin")
}

func TestNoMergeJoinHint(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("create table t1(a int, key(a));")
	tk.MustExec("create table t2(a int, key(a));")
	tk.MustExec("create table t3(a int, key(a));")
	tk.MustExec("create table t4(a int, key(a));")
	runJoinReorderTestData(t, tk, "TestNoMergeJoinHint")
}

func TestNoIndexJoinHint(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec(`set tidb_enable_index_merge_join=true`)
	tk.MustExec("use test")
	tk.MustExec("create table t1(a int, key(a));")
	tk.MustExec("create table t2(a int, key(a));")
	tk.MustExec("create table t3(a int, key(a));")
	tk.MustExec("create table t4(a int, key(a));")
	runJoinReorderTestData(t, tk, "TestNoIndexJoinHint")
}

func TestLeadingJoinHint(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4, t5, t6, t7, t8;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("create table t7(a int, b int, key(a));")
	tk.MustExec("create table t8(a int, b int, key(a));")
	runJoinReorderTestData(t, tk, "TestLeadingJoinHint")

	// test cases for multiple leading hints
	tk.MustExec("select /*+ leading(t1) leading(t2) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use one leading hint at most, when multiple leading hints are used, all leading hints will be invalid"))
}

func TestJoinOrderHint(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4, t5, t6, t7, t8;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("create table t7(a int, b int, key(a));")
	tk.MustExec("create table t8(a int, b int, key(a));")

	// test cases for using the leading hint and straight_join hint at the same time
	tk.MustExec("select /*+ leading(t1) straight_join() */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use the straight_join hint, when we use the leading hint and straight_join hint at the same time, all leading hints will be invalid"))

	tk.MustExec("select /*+ straight_join() leading(t1) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use the straight_join hint, when we use the leading hint and straight_join hint at the same time, all leading hints will be invalid"))

	// more join order hints appear in the same time
	tk.MustExec("select /*+ leading(t1) leading(t1) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use one leading hint at most, when multiple leading hints are used, all leading hints will be invalid"))

	tk.MustExec("select /*+ leading(t1) leading(t2) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use one leading hint at most, when multiple leading hints are used, all leading hints will be invalid"))

	tk.MustExec("select /*+ straight_join() straight_join() */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1105 STRAIGHT_JOIN() is defined more than once, only the last definition takes effect"))

	// test cases for table name in hint
	// the same table appears in the leading hint
	tk.MustExec("select /*+ leading(t1, t1) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t1) in optimizer hint /*+ LEADING(t1, t1) */. Maybe you can use the table alias name",
		"Warning 1815 leading hint is inapplicable, check if the leading hint table is valid"))

	tk.MustExec("select /*+ leading(t1, t2, t1) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t1) in optimizer hint /*+ LEADING(t1, t2, t1) */. Maybe you can use the table alias name",
		"Warning 1815 leading hint is inapplicable, check if the leading hint table is valid"))

	// the wrong table appears in the leading hint
	tk.MustExec("select /*+ leading(t) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t) in optimizer hint /*+ LEADING(t) */. Maybe you can use the table alias name"))

	tk.MustExec("select /*+ leading(t1, t2, t) */ * from t1 join t2 on t1.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t) in optimizer hint /*+ LEADING(t1, t2, t) */. Maybe you can use the table alias name",
		"Warning 1815 leading hint is inapplicable, check if the leading hint table is valid"))

	// table alias in the leading hint
	tk.MustExec("select /*+ leading(t) */ * from t1 t join t2 on t.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows())

	tk.MustExec("select /*+ leading(t1) */ * from t1 t join t2 on t.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t1) in optimizer hint /*+ LEADING(t1) */. Maybe you can use the table alias name"))

	tk.MustExec("select /*+ leading(t2, t) */ * from t1 t join t2 on t.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows())

	tk.MustExec("select /*+ leading(t2, t1) */ * from t1 t join t2 on t.a=t2.a join t3 on t2.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t1) in optimizer hint /*+ LEADING(t2, t1) */. Maybe you can use the table alias name",
		"Warning 1815 leading hint is inapplicable, check if the leading hint table is valid"))

	// table name in leading hint cross query block
	// Todo: Can not handle this case yet. Because when we extract the join group, it will get the join group {t1, t2, t3}.
	// So the table 't4' can not be used.
	tk.MustExec("select /*+ leading(t4) */ * from (select t2.b from t1 join t2 on t1.a=t2.a) t4 join t3 on t4.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 leading hint is inapplicable, check if the leading hint table is valid"))

	tk.MustExec("select /*+ leading(t3, t2@sel_2) */ * from (select t2.b from t1 join t2 on t1.a=t2.a) t4 join t3 on t4.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t2) in optimizer hint /*+ LEADING(t3, t2) */. Maybe you can use the table alias name"))

	tk.MustExec("select * from (select /*+ leading(t1, t3@sel_1) */ t2.b from t1 join t2 on t1.a=t2.a) t4 join t3 on t4.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 There are no matching table names for (t3) in optimizer hint /*+ LEADING(t1, t3) */. Maybe you can use the table alias name"))

	tk.MustExec("select /*+ leading(t3) */ * from (select /*+ leading(t1) */ t2.b from t1 join t2 on t1.a=t2.a) t4 join t3 on t4.b=t3.b")
	tk.MustQuery("show warnings").Check(testkit.Rows("Warning 1815 We can only use one leading hint at most, when multiple leading hints are used, all leading hints will be invalid"))

	runJoinReorderTestData(t, tk, "TestJoinOrderHint")
}

func TestJoinOrderHint4StaticPartitionTable(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3;")
	tk.MustExec(`create table t(a int, b int) partition by hash(a) partitions 3`)
	tk.MustExec(`create table t1(a int, b int) partition by hash(a) partitions 4`)
	tk.MustExec(`create table t2(a int, b int) partition by hash(a) partitions 5`)
	tk.MustExec(`create table t3(a int, b int) partition by hash(b) partitions 3`)
	tk.MustExec(`create table t4(a int, b int) partition by hash(a) partitions 4`)
	tk.MustExec(`create table t5(a int, b int) partition by hash(a) partitions 5`)
	tk.MustExec(`create table t6(a int, b int) partition by hash(b) partitions 3`)

	tk.MustExec(`set @@tidb_partition_prune_mode="static"`)
	tk.MustExec("set @@tidb_enable_outer_join_reorder=true")
	runJoinReorderTestData(t, tk, "TestJoinOrderHint4StaticPartitionTable")
}

func TestJoinOrderHint4DynamicPartitionTable(t *testing.T) {
	failpoint.Enable("github.com/pingcap/tidb/planner/core/forceDynamicPrune", `return(true)`)
	defer failpoint.Disable("github.com/pingcap/tidb/planner/core/forceDynamicPrune")
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t, t1, t2, t3;")
	tk.MustExec(`create table t(a int, b int) partition by hash(a) partitions 3`)
	tk.MustExec(`create table t1(a int, b int) partition by hash(a) partitions 4`)
	tk.MustExec(`create table t2(a int, b int) partition by hash(a) partitions 5`)
	tk.MustExec(`create table t3(a int, b int) partition by hash(b) partitions 3`)
	tk.MustExec(`create table t4(a int, b int) partition by hash(a) partitions 4`)
	tk.MustExec(`create table t5(a int, b int) partition by hash(a) partitions 5`)
	tk.MustExec(`create table t6(a int, b int) partition by hash(b) partitions 3`)

	tk.MustExec(`set @@tidb_partition_prune_mode="dynamic"`)
	tk.MustExec("set @@tidb_enable_outer_join_reorder=true")
	runJoinReorderTestData(t, tk, "TestJoinOrderHint4DynamicPartitionTable")
}

func TestJoinOrderHint4DifferentJoinType(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4, t5, t6, t7, t8;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("create table t7(a int, b int, key(a));")
	tk.MustExec("create table t8(a int, b int, key(a));")
	tk.MustExec("set @@tidb_enable_outer_join_reorder=true")

	runJoinReorderTestData(t, tk, "TestJoinOrderHint4DifferentJoinType")
}

func TestJoinOrderHint4TiFlash(t *testing.T) {
	store := testkit.CreateMockStore(t)
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("drop table if exists t, t1, t2, t3;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("set @@tidb_enable_outer_join_reorder=true")

	// Create virtual tiflash replica info.
	dom := domain.GetDomain(tk.Session())
	is := dom.InfoSchema()
	db, exists := is.SchemaByName(model.NewCIStr("test"))
	require.True(t, exists)
	for _, tblInfo := range db.Tables {
		tableName := tblInfo.Name.L
		if tableName == "t" || tableName == "t1" || tableName == "t2" || tableName == "t3" || tableName == "t4" || tableName == "t5" || tableName == "t6" {
			tblInfo.TiFlashReplica = &model.TiFlashReplicaInfo{
				Count:     1,
				Available: true,
			}
		}
	}

	tk.MustExec("set @@tidb_allow_mpp=1; set @@tidb_enforce_mpp=1;")
	runJoinReorderTestData(t, tk, "TestJoinOrderHint4TiFlash")
}

func TestJoinOrderHint4Subquery(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4, t5, t6, t7, t8;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("create table t7(a int, b int, key(a));")
	tk.MustExec("create table t8(a int, b int, key(a));")
	tk.MustExec("insert into t3 values(1, 1), (2, 2), (3, 3);")
	tk.MustExec("analyze table t3;")

	runJoinReorderTestData(t, tk, "TestJoinOrderHint4Subquery")
}

func TestLeadingJoinHint4OuterJoin(t *testing.T) {
	store := testkit.CreateMockStore(t)

	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test")
	tk.MustExec("set tidb_cost_model_version=2")
	tk.MustExec("drop table if exists t, t1, t2, t3, t4, t5, t6, t7, t8;")
	tk.MustExec("create table t(a int, b int, key(a));")
	tk.MustExec("create table t1(a int, b int, key(a));")
	tk.MustExec("create table t2(a int, b int, key(a));")
	tk.MustExec("create table t3(a int, b int, key(a));")
	tk.MustExec("create table t4(a int, b int, key(a));")
	tk.MustExec("create table t5(a int, b int, key(a));")
	tk.MustExec("create table t6(a int, b int, key(a));")
	tk.MustExec("create table t7(a int, b int, key(a));")
	tk.MustExec("create table t8(a int, b int, key(a));")
	tk.MustExec("set @@tidb_enable_outer_join_reorder=true")
	runJoinReorderTestData(t, tk, "TestLeadingJoinHint4OuterJoin")
}
