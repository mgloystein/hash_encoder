package service_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgloystein/hash_encoder/config"
	"github.com/mgloystein/hash_encoder/service"
)

func TestCreateHash(t *testing.T) {
	c := config.DefaultConfig()

	testObject, err := service.NewHasEncoderService(c)

	if err != nil {
		t.Errorf("Creating the hash encoder service resulted in an error, see below \n %+v", err)
	}

	itemID := testObject.CreateHash("item1")
	_, err = testObject.GetHashedItem(itemID)

	if err == nil {
		t.Errorf("Should not have resulted in an error")
	}

	if err.Error() != fmt.Sprintf("No item found at %d", itemID) {
		t.Errorf(`Expected error to be "No item found at %d"`, itemID)
	}

	time.Sleep((c.WriteDelay + 1) * time.Second)

	itemValue, err := testObject.GetHashedItem(itemID)

	if err != nil {
		t.Errorf("Get resulted in an error, see below \n %+v", err)
	}

	if itemValue != "sXPrOsBtd6oI6KZMpLLQZOMkdJnpjdKGYf9RrLxG0no=" {
		t.Errorf(`Expected %s to mbe "sXPrOsBtd6oI6KZMpLLQZOMkdJnpjdKGYf9RrLxG0no="`, itemValue)
	}
}

func TestStats(t *testing.T) {
	c := config.DefaultConfig()

	testObject, err := service.NewHasEncoderService(c)

	if err != nil {
		t.Errorf("Creating the hash encoder service resulted in an error, see below \n %+v", err)
	}

	_ = testObject.CreateHash("item1")
	_ = testObject.CreateHash("item2")
	_ = testObject.CreateHash("item3")
	_ = testObject.CreateHash("item4")
	_ = testObject.CreateHash("item5")

	time.Sleep((c.WriteDelay + 1) * time.Second)

	stat := testObject.Stats()

	if stat.Count != 5 {
		t.Errorf("Expected 5 items, got %d", stat.Count)
	}

	if stat.AverageProcessTime <= 0.0 {
		t.Errorf("Expected positive average time, got %.2fms", stat.AverageProcessTime)
	}
}
