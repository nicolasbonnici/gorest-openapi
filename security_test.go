package openapi

import "testing"

func reqLen(t *testing.T, op map[string]interface{}, key string) int {
	t.Helper()
	sec, ok := op["security"].([]map[string]interface{})
	if !ok {
		t.Fatalf("%s: operation has no []map security, got %T", key, op["security"])
	}
	return len(sec)
}

func TestOperationSecurity(t *testing.T) {
	tests := []struct {
		path       string
		method     string
		wantSecure bool // true => bearerAuth required, false => explicitly public
	}{
		{"/auth/login", "POST", false},
		{"/auth/register", "POST", false},
		{"/auth/refresh", "POST", false},
		{"/status", "GET", false},
		{"/openapi", "GET", false},
		{"/posts", "GET", false},
		{"/posts", "POST", true},
		{"/comments/:id", "DELETE", true},
		{"/categories/:id/resources", "POST", true},
		{"/metrics", "GET", false},
	}

	for _, tc := range tests {
		got := operationSecurity(tc.path, tc.method)
		secure := len(got) > 0
		if secure != tc.wantSecure {
			t.Errorf("%s %s: wantSecure=%v, got security=%v", tc.method, tc.path, tc.wantSecure, got)
		}
		if secure {
			if _, ok := got[0]["bearerAuth"]; !ok {
				t.Errorf("%s %s: secured requirement should name bearerAuth, got %v", tc.method, tc.path, got)
			}
		}
	}
}

func TestCRUDOperationsCarrySecurity(t *testing.T) {
	res := resourceDTOs{Name: "post", PluralName: "posts"}

	coll := buildCollectionEndpoints(res, "Post", GeneratorConfig{PaginationLimit: 20, PaginationMaxLimit: 100})
	if n := reqLen(t, coll["get"].(map[string]interface{}), "collection get"); n != 0 {
		t.Errorf("collection GET should be public (empty security), got %d requirements", n)
	}
	if n := reqLen(t, coll["post"].(map[string]interface{}), "collection post"); n == 0 {
		t.Error("collection POST should require authentication")
	}

	item := buildItemEndpoints(res, "Post")
	if n := reqLen(t, item["get"].(map[string]interface{}), "item get"); n != 0 {
		t.Errorf("item GET should be public (empty security), got %d requirements", n)
	}
	if n := reqLen(t, item["put"].(map[string]interface{}), "item put"); n == 0 {
		t.Error("item PUT should require authentication")
	}
	if n := reqLen(t, item["delete"].(map[string]interface{}), "item delete"); n == 0 {
		t.Error("item DELETE should require authentication")
	}
}
