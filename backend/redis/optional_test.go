package redis_test

import (
	"testing"

	infrredis "github.com/numduel/numduel/redis"
)

func TestOptionalStoreHelpersReturnNilInterface(t *testing.T) {
	var store *infrredis.Store
	if infrredis.JWTRevoker(store) != nil {
		t.Fatal("JWTRevoker should return nil interface")
	}
	if infrredis.ForceLogoutStore(store) != nil {
		t.Fatal("ForceLogoutStore should return nil interface")
	}
	if infrredis.WSSessionStore(store) != nil {
		t.Fatal("WSSessionStore should return nil interface")
	}
}
