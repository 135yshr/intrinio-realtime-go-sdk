package intriniorealtime

import (
	"strings"
	"testing"
	"time"
)

const (
	yourIntrinioAPIUserName = "YOUR_INTRINIO_API_USERNAME"
	yourIntrinioAPIPassword = "YOUR_INTRINIO_API_PASSWORD"
)

func TestMain(t *testing.T) {
	TestClientConnect(t)
	TestClientJoin(t)
	TestClientLeave(t)
}

func TestClientConnect(t *testing.T) {
	type args struct {
		username string
		password string
		provider provider
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "QUODDに接続してエラーが発生しないこと",
			args: args{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: QUODD,
			},
			wantErr: false,
		},
		{
			name: "IEXに接続してエラーが発生しないこと",
			args: args{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: IEX,
			},
			wantErr: false,
		},
		{
			name: "間違ったユーザーIDが指定されたときにエラーが発生すること",
			args: args{
				username: "Unknown",
				password: yourIntrinioAPIPassword,
				provider: QUODD,
			},
			wantErr: true,
		},
		{
			name: "間違ったパスワードが指定されたときにエラーが発生すること",
			args: args{
				username: yourIntrinioAPIUserName,
				password: "Unknown",
				provider: QUODD,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := New(tt.args.username, tt.args.password, tt.args.provider)
			err := sut.Connect()
			if (err != nil) != tt.wantErr {
				t.Errorf("connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			time.Sleep(5 * time.Second)
			sut.Disconnect()
		})
	}
}

func TestClientJoin(t *testing.T) {
	type fields struct {
		username string
		password string
		provider provider
	}
	type args struct {
		channels []string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantData bool
		wantErr  bool
	}{
		{
			name: "AAPL.NBを指定してApple株の情報をQUODDから取得できること",
			fields: fields{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: QUODD,
			},
			args: args{
				channels: []string{"AAPL.NB"},
			},
			wantData: true,
			wantErr:  false,
		},
		{
			name: "AAPL.NBとMSFT.NBとGE.NBを指定してApple株の情報をQUODDから取得できること",
			fields: fields{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: QUODD,
			},
			args: args{
				channels: []string{"AAPL.NB", "MSFT.NB", "GE.NB"},
			},
			wantData: true,
			wantErr:  false,
		},
		{
			name: "AAPLを指定してApple株の情報をIEXから取得できること",
			fields: fields{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: IEX,
			},
			args: args{
				channels: []string{"AAPL"},
			},
			wantData: true,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readedData := false
			sut := New(tt.fields.username, tt.fields.password, tt.fields.provider)
			sut.OnQuote(func(data map[string]interface{}) {
				readedData = true
			})
			sut.OnError(func(err error) {
				if (err != nil) != tt.wantErr {
					t.Errorf("event error = %v, wantErr %v", err, tt.wantErr)
				}
			})
			err := sut.Connect()
			if (err != nil) != tt.wantErr {
				t.Errorf("connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer sut.Disconnect()
			sut.Join(tt.args.channels...)
			time.Sleep(10 * time.Second)
			sut.LeaveAll()
			if tt.wantData != readedData {
				t.Errorf("%s", "TestClientJoin() data could not be read.")
			}
		})
	}
}

func TestClientLeave(t *testing.T) {
	type fields struct {
		username string
		password string
		provider provider
	}
	type args struct {
		channels []string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantData bool
		wantErr  bool
	}{
		{
			name: "AAPL.NBとMSFT.NBとGE.NBの中からAAPL.NBのみ除外できること",
			fields: fields{
				username: yourIntrinioAPIUserName,
				password: yourIntrinioAPIPassword,
				provider: QUODD,
			},
			args: args{
				channels: []string{"AAPL.NB"},
			},
			wantData: true,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readedData := false
			sut := New(tt.fields.username, tt.fields.password, tt.fields.provider)
			sut.DebugMode = true
			sut.Connect()
			defer sut.Disconnect()
			sut.OnQuote(func(data map[string]interface{}) {
				event := data["event"].(string)
				if event == "info" {
					message := data["data"].(map[string]interface{})["message"].(string)
					readedData = event == "info" && strings.Contains(message, "unsubscribed")
				}
			})
			sut.Join("AAPL.NB", "MSFT.NB", "GE.NB")
			time.Sleep(5 * time.Second)
			sut.Leave("AAPL.NB")
			time.Sleep(5 * time.Second)
			sut.LeaveAll()
			if tt.wantData != readedData {
				t.Errorf("%s", "TestClientLeave() data could not be read.")
			}
		})
	}
}
