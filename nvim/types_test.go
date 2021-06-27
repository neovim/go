package nvim

import (
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level LogLevel
		want  string
	}{
		{
			name:  "Trace",
			level: LogTraceLevel,
			want:  "TraceLevel",
		},
		{
			name:  "Debug",
			level: LogDebugLevel,
			want:  "DebugLevel",
		},
		{
			name:  "Info",
			level: LogInfoLevel,
			want:  "InfoLevel",
		},
		{
			name:  "Warn",
			level: LogWarnLevel,
			want:  "WarnLevel",
		},
		{
			name:  "Error",
			level: LogErrorLevel,
			want:  "ErrorLevel",
		},
		{
			name:  "unkonwn",
			level: LogLevel(-1),
			want:  "unkonwn Level",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.level.String(); got != tt.want {
				t.Fatalf("LogLevel.String() = %v, want %v", tt.want, got)
			}
		})
	}
}
