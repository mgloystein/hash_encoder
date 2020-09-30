package storage_test

import (
	"testing"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/storage"
)

func TestMemoryStore(t *testing.T) {
	item1 := "item1"
	item2 := "item2"

	c := config.DefaultConfig()

	testObject, _ := storage.NewDataStore(c)

	persistable1 := testObject.Reserve()
	result1 := persistable1.ID()
	err1 := persistable1.Persist(item1)

	if err1 != nil {
		t.Errorf("Storing resulted in an error, see below \n %+v", err1)
	}

	if result1 != 1 {
		t.Errorf("Expected item id (%d) to be 1", result1)
	}

	persistable2 := testObject.Reserve()
	result2 := persistable2.ID()
	err2 := persistable2.Persist(item2)

	if err2 != nil {
		t.Errorf("Storing resulted in an error, see below \n %+v", err2)
	}

	if result2 != 2 {
		t.Errorf("Expected item id (%d) to be 2", result2)
	}

	result3, err3 := testObject.Get(1)

	if err3 != nil {
		t.Errorf("Getting resulted in an error, see below \n %+v", err3)
	}

	if result3 != item1 {
		t.Errorf("Expected %s to be %s", result3, item1)
	}

	result4, err4 := testObject.Get(2)

	if err4 != nil {
		t.Errorf("Getting resulted in an error, see below \n %+v", err4)
	}

	if result4 != item2 {
		t.Errorf("Expected %s to be %s", result4, item2)
	}

	_, err5 := testObject.Get(3)

	if err5 == nil {
		t.Errorf("Getting should have resulted in an error")
	}

	if err5.Error() != "No item found at 3" {
		t.Errorf(`Expected error to be "No item found at 3"`)
	}

	persistable3 := testObject.Reserve()
	result5 := persistable3.ID()

	_, err6 := testObject.Get(3)

	if err6 == nil {
		t.Errorf("Getting should have resulted in an error")
	}

	if err6.Error() != "No item found at 3" {
		t.Errorf(`Expected error to be "No item found at 3"`)
	}

	if result5 != 3 {
		t.Errorf("Expected item id (%d) to be 3", result5)
	}
}
