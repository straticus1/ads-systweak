package api

import "testing"

func TestParseKMUtilOutput(t *testing.T) {
	fixture := `No variant specified, falling back to release
Index Refs Address            Size       Wired      Name (Version) UUID <Linked Against>
  123    0 0xffffff7f12340000 0x5000     0x5000     com.vendor.driver (1.2.3) ABCD-1234 <8 6 5>`
	items := parseKMUtilOutput(fixture)
	if len(items) != 1 || items[0].Label != "com.vendor.driver" || items[0].Type != "KernelExtension" {
		t.Fatalf("items = %#v", items)
	}
}

func TestParseSystemExtensionsOutput(t *testing.T) {
	fixture := `2 extension(s)
--- com.apple.system_extension.network_extension (Go to 'System Settings > General > Login Items & Extensions > Network Extensions' to modify these system extension(s))
enabled active teamID bundleID (version) name [state]
* * TEAM123 com.vendor.filter (1.0/42) Vendor Filter [activated enabled]
    TEAM999 com.vendor.pending (2.0/2) Pending Filter [terminated waiting for user]`
	items := parseSystemExtensionsOutput(fixture)
	if len(items) != 2 {
		t.Fatalf("items = %#v", items)
	}
	if items[0].Label != "com.vendor.filter" || !items[0].Enabled || items[0].Type != "SystemExtension" {
		t.Fatalf("first item = %#v", items[0])
	}
	if items[1].Enabled {
		t.Fatalf("pending extension reported enabled: %#v", items[1])
	}
}
