// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fault

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
)

func TestCount(t *testing.T) {
	var ok bool
	var cnt int64
	var ctx = context.TODO()

	Enable()
	AddFaultPoint(ctx, "a", ":5::", "return", 0, "", false)
	AddFaultPoint(ctx, "aa", ":::", "getcount", 0, "a", false)
	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(3), cnt)

	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(4), cnt)

	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(5), cnt)

	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)
	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(6), cnt)

	RemoveFaultPoint(ctx, "a")
	RemoveFaultPoint(ctx, "aa")

	AddFaultPoint(ctx, "a", "3:8:2:", "return", 0, "", false)
	AddFaultPoint(ctx, "aa", ":::", "getcount", 0, "a", false)
	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)
	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(1), cnt)

	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)

	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)
	cnt, _, _ = TriggerFault("aa")
	require.Equal(t, int64(3), cnt)

	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)
	cnt, _, _ = TriggerFault("aa")
	require.Equal(t, int64(4), cnt)

	// 5
	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)

	// 6
	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)

	// 7
	_, _, ok = TriggerFault("a")
	require.Equal(t, true, ok)

	//8
	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)

	//9
	_, _, ok = TriggerFault("a")
	require.Equal(t, false, ok)

	cnt, _, ok = TriggerFault("aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(9), cnt)
	Disable()
	Disable()
}

func wait(t *testing.T) {
	_, _, ok := TriggerFault("w")
	require.Equal(t, true, ok)
}

func TestEcho(t *testing.T) {
	Enable()

	AddFaultPoint(context.TODO(), "e", ":::", "echo", 21, "guns", false)

	i, s, ok := TriggerFault("e")
	require.True(t, ok)
	require.Equal(t, 21, int(i))
	require.Equal(t, "guns", s)

	Disable()
}

func TestWait(t *testing.T) {
	var ok bool
	var cnt int64
	var ctx = context.Background()

	Enable()

	AddFaultPoint(ctx, "w", ":::", "wait", 0, "", false)
	AddFaultPoint(ctx, "n1", ":::", "notify", 0, "w", false)
	AddFaultPoint(ctx, "nall", ":::", "notifyall", 0, "w", false)
	AddFaultPoint(ctx, "gc", ":::", "getcount", 0, "w", false)
	AddFaultPoint(ctx, "gw", ":::", "getwaiters", 0, "w", false)
	AddFaultPoint(ctx, "s", ":::", "sleep", 1, "w", false)

	for i := 0; i < 10; i++ {
		go wait(t)
	}

	TriggerFault("s")

	cnt, _, ok = TriggerFault("gc")
	require.Equal(t, true, ok)
	require.Equal(t, int64(10), cnt)

	cnt, _, ok = TriggerFault("gw")
	require.Equal(t, true, ok)
	require.Equal(t, int64(10), cnt)

	TriggerFault("n1")
	TriggerFault("s")

	cnt, _, ok = TriggerFault("gc")
	require.Equal(t, true, ok)
	require.Equal(t, int64(10), cnt)

	cnt, _, ok = TriggerFault("gw")
	require.Equal(t, true, ok)
	require.Equal(t, int64(9), cnt)

	TriggerFault("nall")
	TriggerFault("s")

	cnt, _, ok = TriggerFault("gc")
	require.Equal(t, true, ok)
	require.Equal(t, int64(10), cnt)

	cnt, _, ok = TriggerFault("gw")
	require.Equal(t, true, ok)
	require.Equal(t, int64(0), cnt)

	Disable()
}

func Test_panic(t *testing.T) {
	var ctx = context.TODO()

	Enable()
	defer Disable()
	require.NoError(t, AddFaultPoint(ctx, "panic_moerr", ":::", "panic", PanicUseMoErr, "use moerr", false))
	require.NoError(t, AddFaultPoint(ctx, "panic_non_moerr", ":::", "panic", PanicUseNonMoErr, "use non moerr", false))

	fun := func(useMoerr bool) {
		defer func() {
			if e := recover(); e != nil {
				_, ok := e.(*moerr.Error)
				if !ok {
					fmt.Println("moerr recovered")
				} else {
					fmt.Println("non moerr recovered")
				}
			}
		}()

		if useMoerr {
			TriggerFault("panic_moerr")
		} else {
			TriggerFault("panic_non_moerr")
		}
	}

	fun(true)
	fun(false)
}

func TestCount2(t *testing.T) {
	var ok bool
	var cnt int64
	var ctx = context.TODO()

	EnableDomain(DomainTest)
	AddFaultPointInDomain(ctx, DomainTest, "a", ":5::", "return", 0, "", false)
	AddFaultPointInDomain(ctx, DomainTest, "aa", ":::", "getcount", 0, "a", false)
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(3), cnt)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(4), cnt)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(5), cnt)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)
	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(6), cnt)

	RemoveFaultPointFromDomain(ctx, DomainTest, "a")
	RemoveFaultPointFromDomain(ctx, DomainTest, "aa")

	AddFaultPointInDomain(ctx, DomainTest, "a", "3:8:2:", "return", 0, "", false)
	AddFaultPointInDomain(ctx, DomainTest, "aa", ":::", "getcount", 0, "a", false)
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)
	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(1), cnt)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)
	cnt, _, _ = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, int64(3), cnt)

	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)
	cnt, _, _ = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, int64(4), cnt)

	// 5
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)

	// 6
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)

	// 7
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, true, ok)

	//8
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)

	//9
	_, _, ok = TriggerFaultInDomain(DomainTest, "a")
	require.Equal(t, false, ok)

	cnt, _, ok = TriggerFaultInDomain(DomainTest, "aa")
	require.Equal(t, true, ok)
	require.Equal(t, int64(9), cnt)
	DisableDomain(DomainTest)
	DisableDomain(DomainTest)
}
