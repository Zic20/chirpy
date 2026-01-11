package auth

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	signature = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
) // VALID ONLY FOR TESTING PURPOSE)

func TestMakeJWT(t *testing.T) {
	id := "09e676d4-76c2-40f9-96ba-f0f4c800b1dd"
	duration := time.Duration.Minutes(5)

	uid, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("could not parse uuid for test case: %s", err.Error())
		return
	}
	token, err := MakeJWT(uid, signature, time.Duration(duration))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(token)
}

func TestValidateJWT(t *testing.T) {
	id := "09e676d4-76c2-40f9-96ba-f0f4c800b1dd"
	duration := 10 * time.Minute

	cases := []struct {
		signature  string
		id         string
		duration   time.Duration
		key        string
		shouldFail bool
	}{
		{
			key:        "With valid data",
			signature:  signature,
			shouldFail: false,
		},
		{
			key:        "with invalid data",
			signature:  "a-string-secret-at-least-256-bits-long",
			shouldFail: true,
		},
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		t.Errorf("could not parse uuid for test case: %s", err.Error())
		return
	}
	token, err := MakeJWT(uid, signature, time.Duration(duration))
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Test case: %v", c.key), func(t *testing.T) {
			userid, err := ValidateJWT(token, c.signature)
			if c.shouldFail {
				if err == nil {
					log.Printf("expected error, got nil (key=%v)", c.key)
					t.Fatalf("expected error, got nil (key=%v)", c.key)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %s", err.Error())
			}
			if uid != userid {
				t.Fatalf("expected id %s, got %s", uid, userid)
			}
		})
	}
}
