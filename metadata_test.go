package hoverfly

import (
	"testing"
)

func TestSetMetadataKey(t *testing.T) {

	metaBucket := GetRandomName(10)
	md := NewBoltDBMetadata(TestDB, metaBucket)

	md.Set("foo", "bar")
	val, err := md.Get("foo")
	expect(t, err, nil)
	expect(t, val, "bar")
}

func TestDeleteMetadataKey(t *testing.T) {
	metaBucket := GetRandomName(10)
	md := NewBoltDBMetadata(TestDB, metaBucket)

	md.Set("foo", "bar")
	md.Delete("foo")

	_, err := md.Get("foo")
	refute(t, err, nil)
}

func TestGetAllValues(t *testing.T) {
	metaBucket := GetRandomName(10)
	md := NewBoltDBMetadata(TestDB, metaBucket)

	md.Set("foo", "bar")
	md.Set("foo2", "bar2")
	md.Set("foo3", "bar3")

	values, err := md.GetAll()
	expect(t, err, nil)
	expect(t, len(values), 3)
	expect(t, values["foo"], "bar")
}

func TestDeleteAllData(t *testing.T) {
	metaBucket := GetRandomName(10)
	md := NewBoltDBMetadata(TestDB, metaBucket)

	md.Set("foo", "bar")
	md.Set("foo2", "bar2")
	md.Set("foo3", "bar3")

	md.DeleteData()

	values, err := md.GetAll()
	expect(t, err, nil)
	expect(t, len(values), 0)
}
