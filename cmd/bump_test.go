package cmd_test

import (
	"reflect"
	"testing"

	"github.com/chanzuckerberg/bff/cmd"

	"github.com/blang/semver"
)

func TestReleaseType(t *testing.T) {
	type args struct {
		major    uint64
		breaking bool
		feature  bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"pre1-patch", args{0, false, false}, "patch"},
		{"pre1-patch", args{0, true, false}, "minor"},
		{"pre1-patch", args{0, false, true}, "minor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.ReleaseType(tt.args.major, tt.args.breaking, tt.args.feature); got != tt.want {
				t.Errorf("releaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewVersion(t *testing.T) {
	type args struct {
		ver         semver.Version
		releaseType string
	}
	tests := []struct {
		name string
		args args
		want semver.Version
	}{
		{"patch", args{semver.MustParse("0.1.0"), "patch"}, semver.MustParse("0.1.1")},
		{"minor", args{semver.MustParse("0.0.0"), "minor"}, semver.MustParse("0.1.0")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmd.NewVersion(tt.args.ver, tt.args.releaseType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
