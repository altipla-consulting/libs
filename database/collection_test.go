package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"libs.altipla.consulting/errors"
)

func TestGet(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	require.NoError(t, testDB.Exec(ctx, `INSERT INTO testing(code, name, revision) VALUES ("foo", "foov", 1), ("bar", "barv", 2)`))

	m := &testingModel{
		Code: "bar",
	}
	require.NoError(t, testings.Get(ctx, m))

	require.Equal(t, "barv", m.Name)
	require.EqualValues(t, 2, m.Tracking().StoredRevision())
}

func TestGetNotFound(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "untouch",
	}
	require.EqualError(t, testings.Get(ctx, m), ErrNoSuchEntity.Error())
}

func TestGetNotTouchCols(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "untouched",
	}
	require.EqualError(t, testings.Get(ctx, m), ErrNoSuchEntity.Error())

	require.Equal(t, "untouched", m.Name)
}

func TestInsert(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, other))
	require.Equal(t, "bar", other.Name)
	require.EqualValues(t, 0, other.Tracking().StoredRevision())
}

func TestInsertOnBeforePutHooker(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingHooker{
		Code: "foo",
	}

	require.NoError(t, testingsHooker.Put(ctx, m))

	other := &testingHooker{
		Code: "foo",
	}

	require.NoError(t, testingsHooker.Get(ctx, other))
	require.Equal(t, other.Changed, "changed")
}

func TestInsertOnAfterPutHooker(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingHooker{
		Code: "foo",
	}

	require.NoError(t, testingsHooker.Put(ctx, m))

	other := &testingHooker{
		Code: "foo",
	}

	require.NoError(t, testingsHooker.Get(ctx, other))
	require.True(t, m.Executed)
	require.False(t, other.Executed)
}

func TestInsertAuto(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))
	require.EqualValues(t, m.ID, 1)

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))
	require.EqualValues(t, m.ID, 2)

	other := &testingAutoModel{
		ID: 1,
	}
	require.NoError(t, testingsAuto.Get(ctx, other))
	require.Equal(t, "foo", other.Name)

	other = &testingAutoModel{
		ID: 2,
	}
	require.NoError(t, testingsAuto.Get(ctx, other))
	require.Equal(t, "bar", other.Name)
}

func TestUpdate(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.NoError(t, testings.Put(ctx, m))
	require.NoError(t, testings.Get(ctx, m))

	m.Name = "qux"
	require.NoError(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.NoError(t, testings.Get(ctx, other))
	require.Equal(t, "qux", other.Name)
	require.EqualValues(t, 1, other.Tracking().StoredRevision())
}

func TestUpdateConcurrentTransaction(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.NoError(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.NoError(t, testings.Get(ctx, other))
	other.Name = "baz"
	require.NoError(t, testings.Put(ctx, other))

	m.Name = "qux"
	require.EqualError(t, testings.Put(ctx, m), ErrConcurrentTransaction.Error())

	check := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, check))
	require.Equal(t, "baz", check.Name)
}

func TestInsertAndUpdate(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	m.Name = "qux"
	require.Nil(t, testings.Put(ctx, m))

	other := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Get(ctx, other))
	require.Equal(t, "qux", other.Name)
	require.EqualValues(t, 1, other.Tracking().StoredRevision())
}

func TestTransactionalPut(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	require.Nil(t, testings.Put(ctx, &testingModel{Code: "foo"}))

	count, err := testings.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, count, 1)

	fn := func(ctx context.Context) error {
		require.Nil(t, testings.Put(ctx, &testingModel{Code: "bar"}))

		// Explicit context.Background() to test as a different client.
		count, err := testings.Count(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, count, 1)

		return nil
	}
	require.NoError(t, testDB.RunTransaction(ctx, fn))

	count, err = testings.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, count, 2)
}

func TestTransactionalPutRollback(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	require.Nil(t, testings.Put(ctx, &testingModel{Code: "foo"}))

	count, err := testings.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, count, 1)

	fn := func(ctx context.Context) error {
		require.Nil(t, testings.Put(ctx, &testingModel{Code: "bar"}))

		// Explicit context.Background() to test as a different client.
		count, err := testings.Count(context.Background())
		require.NoError(t, err)
		require.EqualValues(t, count, 1)

		return errors.Errorf("foo")
	}
	require.EqualError(t, testDB.RunTransaction(ctx, fn), "foo")

	count, err = testings.Count(ctx)
	require.NoError(t, err)
	require.EqualValues(t, count, 1)
}

func TestDelete(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	n, err := testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 1)

	require.Nil(t, testings.Delete(ctx, m))

	n, err = testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 0)
}

func TestGetAllEmpty(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	var models []*testingModel
	require.Nil(t, testings.GetAll(ctx, &models))

	require.Len(t, models, 0)
}

func TestGetAll(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "foo name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "bar name",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "bar name", models[0].Name)
	require.EqualValues(t, 0, models[0].Tracking().StoredRevision())

	require.Equal(t, "foo", models[1].Code)
	require.Equal(t, "foo name", models[1].Name)
	require.EqualValues(t, 0, models[1].Tracking().StoredRevision())
}

func TestGetAllFiltering(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "test",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "test",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "qux",
		Name: "ignore",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Filter("name", "test").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "test", models[0].Name)
	require.Equal(t, "foo", models[1].Code)
	require.Equal(t, "test", models[1].Name)
}

func TestGetAllOperator(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "ignore",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("id <=", 2).GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "foo", models[0].Name)
	require.Equal(t, "bar", models[1].Name)
}

func TestGetAllOperatorAndPlaceholder(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "baz",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("name LIKE ?", "ba%").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Name)
	require.Equal(t, "baz", models[1].Name)
}

func TestGetAllOperatorIN(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "foo name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
		Name: "bar name",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "qux",
		Name: "ignore",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Filter("name IN", []string{"foo name", "bar name"}).GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "bar", models[0].Code)
	require.Equal(t, "foo", models[1].Code)
}

func TestGetOrder(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.Order("-code").GetAll(ctx, &models))

	require.Len(t, models, 2)

	require.Equal(t, "foo", models[0].Code)
	require.Equal(t, "bar", models[1].Code)
}

func TestMultipleOrderPanics(t *testing.T) {
	require.PanicsWithValue(t, "call Order multiple times, do not pass multiple columns", func() {
		testings.Order("foo, -bar")
	})
}

func TestAscOrderPanics(t *testing.T) {
	require.PanicsWithValue(t, "do not call Order with `foo ASC`, use plain `foo` instead", func() {
		testings.Order("foo ASC")
	})
}

func TestDescOrderPanics(t *testing.T) {
	require.PanicsWithValue(t, "do not call Order with `foo DESC`, use plain `-foo` instead", func() {
		testings.Order("foo DESC")
	})
}

func TestMultipleFilters(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "qux",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Filter("id >", 1).Filter("id <", 3).GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.Equal(t, "bar", models[0].Name)
}

func TestCount(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	n, err := testingsAuto.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 3)
}

func TestCountFilter(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = new(testingAutoModel)
	require.Nil(t, testingsAuto.Put(ctx, m))

	n, err := testingsAuto.Filter("id >=", 2).Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 2)
}

func TestLimit(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "baz",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.Limit(1).Offset(1).GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.Equal(t, models[0].Name, "bar")
}

func TestGetMultiStrings(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "baz",
	}
	require.Nil(t, testings.Put(ctx, m))

	var models []*testingModel
	require.Nil(t, testings.GetMulti(ctx, []string{"foo", "bar"}, &models))

	require.Len(t, models, 2)

	require.Equal(t, models[0].Code, "foo")
	require.EqualValues(t, 0, models[0].Tracking().StoredRevision())

	require.Equal(t, models[1].Code, "bar")
	require.EqualValues(t, 0, models[1].Tracking().StoredRevision())
}

func TestGetMultiIntegers(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "baz",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.GetMulti(ctx, []int64{3, 2}, &models))

	require.Len(t, models, 2)

	require.Equal(t, models[0].Name, "baz")
	require.Equal(t, models[1].Name, "bar")
}

func TestGetMultiError(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	err := testingsAuto.GetMulti(ctx, []int64{2, 1}, &models)
	require.EqualError(t, err, "no such entity; <nil>")

	merr, ok := err.(MultiError)
	require.True(t, ok)
	require.EqualError(t, merr[0], ErrNoSuchEntity.Error())
	require.Nil(t, merr[1])

	require.Len(t, models, 2)

	require.Nil(t, models[0])
	require.Equal(t, models[1].Name, "foo")
}

func TestGetMultiEmpty(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	var models []*testingModel
	require.Nil(t, testings.GetMulti(ctx, []string{}, &models))
}

func TestFirst(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	require.Nil(t, testDB.Exec(ctx, `INSERT INTO testing(code, name, revision) VALUES ("foo", "foov", 1), ("bar", "barv", 2)`))

	m := new(testingModel)
	require.Nil(t, testings.Filter("code", "bar").First(ctx, m))

	require.Equal(t, "barv", m.Name)
	require.EqualValues(t, 2, m.Tracking().StoredRevision())
}

func TestFirstNotFound(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := new(testingModel)
	require.EqualError(t, testings.Filter("code", "foo").First(ctx, m), ErrNoSuchEntity.Error())
}

func TestFirstNotTouchCols(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Name: "untouched",
	}
	require.EqualError(t, testings.Filter("code", "foo").First(ctx, m), ErrNoSuchEntity.Error())

	require.Equal(t, "untouched", m.Name)
}

func TestTruncate(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingModel{
		Code: "foo",
		Name: "bar",
	}
	require.Nil(t, testings.Put(ctx, m))

	m = &testingModel{
		Code: "baz",
		Name: "qux",
	}
	require.Nil(t, testings.Put(ctx, m))

	n, err := testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 2)

	require.Nil(t, testings.Truncate(ctx))

	n, err = testings.Count(ctx)
	require.Nil(t, err)
	require.EqualValues(t, n, 0)
}

func TestTruncateResetAutoIncrement(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	m := &testingAutoModel{
		Name: "bar",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	m = &testingAutoModel{
		Name: "qux",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	require.Nil(t, testingsAuto.Truncate(ctx))

	m = &testingAutoModel{
		Name: "foo",
	}
	require.Nil(t, testingsAuto.Put(ctx, m))

	var models []*testingAutoModel
	require.Nil(t, testingsAuto.GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterExists(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	child := &testingRelChild{
		Parent: parent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	var models []*testingRelParent
	require.Nil(t, testingsRelParent.FilterExists(testingsRelChild.Filter("foo", "foo-value"), "testing_relparent.id = testing_relchild.parent").GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterExistsDoesNotAffectSubquery(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	child := &testingRelChild{
		Parent: parent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	var unused []*testingRelParent
	subquery := testingsRelChild.Filter("foo", "foo-value")
	require.Nil(t, testingsRelParent.FilterExists(subquery, "testing_relparent.id = testing_relchild.parent").GetAll(ctx, &unused))

	var submodels []*testingRelChild
	require.Nil(t, subquery.GetAll(ctx, &submodels))

	require.Len(t, submodels, 1)

	require.EqualValues(t, submodels[0].ID, 1)
}

func TestFilterExistsAliases(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	child := &testingRelChild{
		Parent: parent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	var models []*testingRelParent
	require.Nil(t, testingsRelParent.Alias("alias1").FilterExists(testingsRelChild.Filter("foo", "foo-value").Alias("alias2"), "alias1.id = alias2.parent").GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterExistsAliasesAfterTheFilter(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	child := &testingRelChild{
		Parent: parent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	var models []*testingRelParent
	require.Nil(t, testingsRelParent.FilterExists(testingsRelChild.Filter("foo", "foo-value").Alias("alias2"), "alias1.id = alias2.parent").Alias("alias1").GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterNotExists(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	child := &testingRelChild{
		Parent: otherParent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	var models []*testingRelParent
	require.Nil(t, testingsRelParent.FilterNotExists(testingsRelChild.Filter("foo", "foo-value"), "testing_relparent.id = testing_relchild.parent").GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterNotExistsDoesNotAffectSubquery(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	child := &testingRelChild{
		Parent: parent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	var unused []*testingRelParent
	subquery := testingsRelChild.Filter("foo", "foo-value")
	require.Nil(t, testingsRelParent.FilterNotExists(subquery, "testing_relparent.id = testing_relchild.parent").GetAll(ctx, &unused))

	var submodels []*testingRelChild
	require.Nil(t, subquery.GetAll(ctx, &submodels))

	require.Len(t, submodels, 1)

	require.EqualValues(t, submodels[0].ID, 1)
}

func TestFilterNotExistsAliases(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	child := &testingRelChild{
		Parent: otherParent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	var models []*testingRelParent
	q := testingsRelParent.Alias("alias1").FilterNotExists(testingsRelChild.Filter("foo", "foo-value").Alias("alias2"), "alias1.id = alias2.parent")
	require.Nil(t, q.GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}

func TestFilterNotExistsAliasesAfterTheFilter(t *testing.T) {
	initDatabase(t)
	defer closeDatabase(t)
	ctx := context.Background()

	parent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, parent))

	otherParent := new(testingRelParent)
	require.Nil(t, testingsRelParent.Put(ctx, otherParent))

	child := &testingRelChild{
		Parent: otherParent.ID,
		Foo:    "foo-value",
	}
	require.Nil(t, testingsRelChild.Put(ctx, child))

	var models []*testingRelParent
	q := testingsRelParent.FilterNotExists(testingsRelChild.Filter("foo", "foo-value").Alias("alias2"), "alias1.id = alias2.parent").Alias("alias1")
	require.Nil(t, q.GetAll(ctx, &models))

	require.Len(t, models, 1)

	require.EqualValues(t, models[0].ID, 1)
}
