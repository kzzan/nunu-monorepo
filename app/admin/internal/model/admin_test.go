// model 包的领域逻辑单元测试。
package model

import "testing"

// TestParsePermissionKeyUsesLastSeparator 验证含逗号资源的键解析
// 以最后一个分隔符为准。
func TestParsePermissionKeyUsesLastSeparator(t *testing.T) {
	permission, ok := ParsePermissionKey("api:/v1/items,a,GET")
	if !ok {
		t.Fatal("ParsePermissionKey() rejected valid key")
	}
	if permission.Resource != "api:/v1/items,a" || permission.Action != "GET" {
		t.Fatalf("ParsePermissionKey() = %#v", permission)
	}
}

// TestParsePermissionKeyRejectsMalformedValue 验证各类非法键被拒绝。
func TestParsePermissionKeyRejectsMalformedValue(t *testing.T) {
	for _, key := range []string{"", "api:/v1/items", ",GET", "api:/v1/items,"} {
		if _, ok := ParsePermissionKey(key); ok {
			t.Fatalf("ParsePermissionKey(%q) expected rejection", key)
		}
	}
}
