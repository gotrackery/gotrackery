package traccar

import (
	"encoding/json"
	"testing"

	"github.com/gotrackery/protocol"
)

func TestAttributes(t *testing.T) {
	type args struct {
		a protocol.Attributes
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic test",
			args: args{a: protocol.Attributes{"i": 1, "f": 2.1, "s": "3"}},
			want: `{"f":2.1,"i":1,"s":"3"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.args.a)
			if err != nil {
				t.Error(err)
			}
			if string(got) != tt.want {
				t.Errorf("Attributes() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
