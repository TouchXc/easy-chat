package logic

import (
	"context"
	"easy-chat/apps/user/rpc/user"
	"reflect"
	"testing"
)

func TestRegisterLogic_Register(t *testing.T) {
	type args struct {
		in *user.RegisterReq
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"1", args{in: &user.RegisterReq{
				Phone:    "13715781956",
				Nickname: "xjs",
				Password: "123456",
				Avatar:   "png.jpg",
				Sex:      1,
			}}, true, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewRegisterLogic(context.Background(), svcCtx)
			got, err := l.Register(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Log(tt.name, got)
			}
		})
	}
}
