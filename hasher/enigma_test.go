package hasher_test

import (
	"testing"

	"github.com/mgloystein/hash_encoder/hasher"
)

type pair struct {
	a, b string
}

const (
	secret = "imarealtivelylongandsomewhatsecuresecret"
)

var (
	tests = []pair{
		{"testing1", "MD0pnmIWLSKUz0FoyUl7HFXU8DVyq+L4j8JkMg7p0uc="},
	}
)

func TestHash(t *testing.T) {
	testObject, _ := hasher.NewEnigma(secret)

	result, err := testObject.Generate(tests[0].a)

	if err != nil {
		t.Errorf("Hash resulting in an error, see below \n %+v", err)
	}

	if result != tests[0].b {
		t.Errorf("Expected %s to equal %s", result, tests[0].b)
	}
}

func TestSecretLengthError(t *testing.T) {
	_, err := hasher.NewEnigma("secret")

	if err == nil {
		t.Errorf("An error was expected but not returned")
	}

	if err.Error() != "Secret is invalid, it should be at least 32 charaters" {
		t.Error("Expected error to equal \"Secret is invalid, it should be at least 32 charaters\"")
	}
}
