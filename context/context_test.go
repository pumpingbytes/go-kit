package context

import (
	"bytes"
	stdcontext "context"
	"reflect"
	"testing"

	"github.com/ygrebnov/keys"
)

type testValueCtxKey struct{}

func TestCtx(t *testing.T) {
	tests := []struct {
		name string
		key  keys.Key
		val  any
		rest []any
		want Context
	}{
		{
			name: "single pair",
			key:  keys.Key("k"),
			val:  1,
			want: Context{keys.Key("k"): 1},
		},
		{
			name: "additional pairs",
			key:  keys.Key("a"),
			val:  1,
			rest: []any{keys.Key("b"), 2, keys.Key("c"), 3},
			want: Context{keys.Key("a"): 1, keys.Key("b"): 2, keys.Key("c"): 3},
		},
		{
			name: "odd rest ignores trailing",
			key:  keys.Key("a"),
			val:  1,
			rest: []any{keys.Key("b"), 2, keys.Key("c")},
			want: Context{keys.Key("a"): 1, keys.Key("b"): 2},
		},
		{
			name: "non-keys.Key keys are ignored",
			key:  keys.Key("a"),
			val:  1,
			rest: []any{123, "nope", true, "also nope", keys.Key("b"), 2},
			want: Context{keys.Key("a"): 1, keys.Key("b"): 2},
		},
		{
			name: "later values overwrite earlier",
			key:  keys.Key("a"),
			val:  1,
			rest: []any{keys.Key("a"), 2, keys.Key("a"), 3},
			want: Context{keys.Key("a"): 3},
		},
		{
			name: "rest can set initial key to nil",
			key:  keys.Key("a"),
			val:  1,
			rest: []any{keys.Key("a"), nil},
			want: Context{keys.Key("a"): nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := Ctx(tt.key, tt.val, tt.rest...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Ctx() mismatch\n got: %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}

func TestContextMarshal(t *testing.T) {
	tests := []struct {
		name    string
		ctx     Context
		want    []byte
		wantErr bool
	}{
		{
			name:    "valid context",
			ctx:     Context{keys.Key("request.id"): "r1", keys.Key("trace.id"): "t1"},
			want:    []byte(`{"request.id":"r1","trace.id":"t1"}`),
			wantErr: false,
		},
		{
			name:    "marshal error",
			ctx:     Context{keys.Key("bad"): make(chan int)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ctx.Marshal()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Fatalf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPutValueAndGetValue(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		ctx := PutValue(stdcontext.Background(), testValueCtxKey{}, 42)
		if got := GetValue(ctx, testValueCtxKey{}, 0); got != 42 {
			t.Fatalf("GetValue() = %d, want 42", got)
		}
	})

	t.Run("fallback when missing", func(t *testing.T) {
		if got := GetValue(stdcontext.Background(), testValueCtxKey{}, "fallback"); got != "fallback" {
			t.Fatalf("GetValue() = %q, want fallback", got)
		}
	})

	t.Run("fallback on wrong type", func(t *testing.T) {
		ctx := stdcontext.WithValue(stdcontext.Background(), testValueCtxKey{}, "not-an-int")
		if got := GetValue(ctx, testValueCtxKey{}, 7); got != 7 {
			t.Fatalf("GetValue() = %d, want 7", got)
		}
	})

	t.Run("typed nil pointer is preserved", func(t *testing.T) {
		ctx := PutValue[*bytes.Buffer](stdcontext.Background(), testValueCtxKey{}, nil)
		if got := GetValue[*bytes.Buffer](ctx, testValueCtxKey{}, bytes.NewBufferString("fallback")); got != nil {
			t.Fatalf("GetValue() = %v, want nil", got)
		}
	})
}

