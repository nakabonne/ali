package gui

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/nakabonne/ali/attacker"
)

func TestKeybinds(t *testing.T) {
	tests := []struct {
		name string
		key  keyboard.Key
	}{
		{
			name: "quit",
			key:  keyboard.KeyCtrlC,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			go func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						return
					}
				}
			}(ctx)
			f := keybinds(ctx, cancel, nil, "", attacker.Options{})
			f(&terminalapi.Keyboard{Key: tt.key})
			// If ctx wasn't expired, goleak will find it.
		})
	}
}

func TestAttack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name         string
		chartDrawing bool
	}{
		{
			name:         "chart is drawing",
			chartDrawing: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &drawer{
				widgets: &widgets{
					latencyChart:  nil,
					messageText:   nil,
					latenciesText: nil,
					bytesText:     nil,
					othersText:    nil,
					progressGauge: nil,
					navi:          nil,
				},
				messageCh:    make(chan string, 100),
				chartDrawing: tt.chartDrawing,
			}
			attack(ctx, d, "", attacker.Options{})
		})
	}
}

/*func TestMakeOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name    string
		widgets widgets
		want    attacker.Options
		wantErr bool
	}{
		{
			name: "wrong rate limit given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong duration given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong timeout given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong method given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				methodInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong header given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				methodInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				headerInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong body file path given",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				methodInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				headerInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				bodyInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("a")
					return m
				}(),
			},
			wantErr: true,
		},
		{
			name: "default values applied",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				methodInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				headerInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
				bodyInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("")
					return m
				}(),
			},
			want: attacker.Options{
				Rate:     50,
				Duration: 10 * time.Second,
				Timeout:  30 * time.Second,
				Method:   "",
				Header:   http.Header{},
			},
			wantErr: false,
		},
		{
			name: "given values applied",
			widgets: widgets{
				rateLimitInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("1")
					return m
				}(),
				durationInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("1s")
					return m
				}(),
				timeoutInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("1s")
					return m
				}(),
				methodInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("POST")
					return m
				}(),
				headerInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("Foo: Bar")
					return m
				}(),
				bodyInput: func() TextInput {
					m := NewMockTextInput(ctrl)
					m.EXPECT().Read().Return("./testdata/body-1.json")
					return m
				}(),
			},
			want: attacker.Options{
				Rate:     1,
				Duration: time.Second,
				Timeout:  time.Second,
				Method:   "POST",
				Header: http.Header{
					"Foo": []string{
						"Bar",
					},
				},
				Body: []byte(`{"foo": 1}`),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeOptions(&tt.widgets)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
*/
