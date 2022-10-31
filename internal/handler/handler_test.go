package handler

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shortingURL(t *testing.T) {
	type args struct {
		longURL string
	}
	tests := []struct {
		name       string
		args       args
		wantRegexp string
		wantErr    bool
	}{
		{
			name:       "noURL",
			args:       args{longURL: "longlonglonglogn.lg/"},
			wantRegexp: "",
			wantErr:    true,
		},
		{
			name:       "normal URL",
			args:       args{longURL: "http://longlonglonglogn.lg/"},
			wantRegexp: "\\w{8}",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shortingURL(tt.args.longURL)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			assert.Regexp(t, tt.wantRegexp, got)
		})
	}
}
