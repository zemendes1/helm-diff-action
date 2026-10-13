package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostPRComment(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{name: "created", status: http.StatusCreated, wantErr: false},
		{name: "forbidden", status: http.StatusForbidden, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *http.Request
			var gotBody []byte
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r
				gotBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			err := postPRComment("hello", 42, server.URL, "owner/repo", "secret")
			if (err != nil) != tt.wantErr {
				t.Fatalf("postPRComment() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got.Method != http.MethodPost {
				t.Errorf("method = %q, want %q", got.Method, http.MethodPost)
			}
			if want := "/repos/owner/repo/issues/42/comments"; got.URL.Path != want {
				t.Errorf("path = %q, want %q", got.URL.Path, want)
			}
			for header, want := range map[string]string{
				"Authorization": "Bearer secret",
				"Accept":        "application/vnd.github+json",
				"Content-Type":  "application/json",
			} {
				if got.Header.Get(header) != want {
					t.Errorf("%s = %q, want %q", header, got.Header.Get(header), want)
				}
			}
			var body prCommentBody
			if err := json.Unmarshal(gotBody, &body); err != nil {
				t.Fatalf("body is not JSON: %v: %s", err, gotBody)
			}
			if body.Body != "hello" {
				t.Errorf("body = %q, want %q", body.Body, "hello")
			}
		})
	}
}
