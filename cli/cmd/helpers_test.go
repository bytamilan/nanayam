package cmd

import (
	"strings"
	"testing"
)

// formatArgs renders chaincode arguments into the JSON array Fabric expects,
// so quoting has to survive values that contain quotes or spaces.
func TestFormatArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"empty", nil, ""},
		{"single", []string{"InitLedger"}, `"InitLedger"`},
		{"several", []string{"CreateAsset", "asset1", "blue"}, `"CreateAsset","asset1","blue"`},
		{"value with a space", []string{"Create", "blue asset"}, `"Create","blue asset"`},
		{"value with a quote", []string{`say "hi"`}, `"say \"hi\""`},
		{"empty value", []string{"Fn", ""}, `"Fn",""`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatArgs(tc.args); got != tc.want {
				t.Fatalf("formatArgs(%q) = %s, want %s", tc.args, got, tc.want)
			}
		})
	}
}

func TestOrgLowerNormalisesCase(t *testing.T) {
	cases := map[string]string{
		"Org1":      "org1",
		"ACB":       "acb",
		"Dept":      "dept",
		"OVERSIGHT": "oversight",
		"judiciary": "judiciary",
		"":          "",
	}

	for in, want := range cases {
		if got := orgLower(in); got != want {
			t.Errorf("orgLower(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInstallHintForPkgManagerNamesThePackage(t *testing.T) {
	// Whatever the platform, the hint has to be a command the user can act on
	// and must mention the package being installed.
	hint := installHintForPkgManager("jq")

	if hint == "" {
		t.Fatal("installHintForPkgManager() returned an empty hint")
	}
	if !strings.Contains(hint, "jq") {
		t.Errorf("hint %q does not name the package", hint)
	}
}
