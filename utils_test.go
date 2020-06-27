package main

import (
	"reflect"
	"testing"
)

func TestParseRemoteString(t *testing.T) {
	gotUser, gotHost, gotRemoteResource := ParseRemoteString("username@destination_host:source")
	wantUser, wantHost, wantRemoteResource := "username", "destination_host", "source"

	if !reflect.DeepEqual(gotUser, wantUser) {
		t.Fatalf("expected: %v, got: %v", wantUser, gotUser)
	}

	if !reflect.DeepEqual(gotHost, wantHost) {
		t.Fatalf("expected: %v, got: %v", gotHost, wantHost)
	}

	if !reflect.DeepEqual(gotRemoteResource, wantRemoteResource) {
		t.Fatalf("expected: %v, got: %v", gotRemoteResource, wantRemoteResource)
	}
}
