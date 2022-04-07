package app

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestApp_SendAPUs(t *testing.T) {
	type fields struct {
		appx App
	}
	type args struct {
		nameReader   string
		sessionId    string
		closeSession bool
		data         [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				appx: Instance(),
			},
			args: args{
				nameReader: func() string {
					if rds, err := Instance().ListReaders(); err == nil {
						for _, r := range rds {
							if strings.ContainsAny(r, "PICC") {
								t.Logf("reader name: %q", r)
								return r
							}
						}
					}
					return ""
				}(),
				sessionId:    fmt.Sprintf("%d", time.Now().UnixNano()),
				closeSession: true,
				data: [][]byte{
					{0xFF, 0xCA, 0x00, 0x00, 0x00},
					{0xFF, 0xCA, 0x01, 0x00, 0x00},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.fields.appx.SendAPUs(tt.args.nameReader, tt.args.sessionId, tt.args.closeSession, tt.args.data...)
			if (err != nil) != tt.wantErr {
				t.Errorf("App.SendAPUs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			i := 0
			for v := range got {
				t.Logf("App.SendAPUs() = [ %X ], want [ %X ]", tt.args.data[i], v)
				i++
			}
		})
	}
}

func Test_app_VerifyCardInReader(t *testing.T) {
	type fields struct {
		appx App
	}
	type args struct {
		nameReader string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			fields: fields{
				appx: Instance(),
			},
			args: args{
				nameReader: func() string {
					if rds, err := Instance().ListReaders(); err == nil {
						for _, r := range rds {
							if strings.Contains(r, "PICC") {
								// t.Logf("reader name AAAA: %q", r)
								return r
							}
						}
					}
					return ""
				}(),
			},
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				appx: Instance(),
			},
			args: args{
				nameReader: func() string {
					if rds, err := Instance().ListReaders(); err == nil {
						for _, r := range rds {
							if strings.Contains(r, "SAM") {

								return r
							}
						}
					}
					return ""
				}(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("reader name: %q", tt.args.nameReader)
			got, err := tt.fields.appx.VerifyCardInReader(tt.args.nameReader)
			if (err != nil) != tt.wantErr {
				t.Errorf("app.VerifyCardInReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("app.VerifyCardInReader() = %s", got)
			t.Logf("card.Status = %+v", func() interface{} { s, _ := got.Status(); return s }())
			t.Logf("card.Atr = %X", func() interface{} { s, _ := got.Atr(); return s }())
		})
	}
}
