package commands

import "testing"

func TestSplitRankingCommandArgsAllowsInlineRankingArguments(t *testing.T) {
	cases := []struct {
		name      string
		message   string
		command   string
		wantArgs  string
		wantMatch bool
	}{
		{name: "space", message: "sk 100", command: "sk", wantArgs: "100", wantMatch: true},
		{name: "inline rank", message: "sk100", command: "sk", wantArgs: "100", wantMatch: true},
		{name: "inline range", message: "sk1-10", command: "sk", wantArgs: "1-10", wantMatch: true},
		{name: "inline shorthand", message: "cf1k", command: "cf", wantArgs: "1k", wantMatch: true},
		{name: "inline uid", message: "csb123456789", command: "csb", wantArgs: "123456789", wantMatch: true},
		{name: "skip forecast alias", message: "skp", command: "sk", wantMatch: false},
		{name: "skip skline", message: "sk线", command: "sk", wantMatch: false},
		{name: "world link region prefix", message: "cnwlsk1 100", command: "cnwlsk", wantArgs: "1 100", wantMatch: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args, ok := splitRankingCommandArgs(tc.message, tc.command, true)
			if ok != tc.wantMatch || args != tc.wantArgs {
				t.Fatalf("splitRankingCommandArgs(%q, %q) = (%q, %v), want (%q, %v)", tc.message, tc.command, args, ok, tc.wantArgs, tc.wantMatch)
			}
		})
	}
}
