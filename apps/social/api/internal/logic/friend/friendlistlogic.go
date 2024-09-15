package friend

import (
	"context"
	"easy-chat/apps/social/api/internal/svc"
	"easy-chat/apps/social/api/internal/types"
	"easy-chat/apps/social/rpc/socialclient"
	"easy-chat/apps/user/rpc/userclient"
	"easy-chat/pkg/ctxdata"
	"github.com/zeromicro/go-zero/core/logx"
)

type FriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendListLogic {
	return &FriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

//获取好友列表

func (l *FriendListLogic) FriendList(req *types.FriendListReq) (resp *types.FriendListResp, err error) {
	//先从userRpc服务中获取用户信息
	uid := ctxdata.GetUid(l.ctx)
	friends, err := l.svcCtx.Social.FriendList(l.ctx, &socialclient.FriendListReq{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}
	if len(friends.List) == 0 {
		return &types.FriendListResp{}, nil
	}
	//根据好友id查好友信息
	uids := make([]string, 0, len(friends.List))
	for _, v := range friends.List {
		uids = append(uids, v.FriendUid)
	}

	//根据uids查询用户信息
	users, err := l.svcCtx.User.FindUser(l.ctx, &userclient.FindUserReq{
		Ids: uids,
	})
	if err != nil {
		return &types.FriendListResp{}, nil
	}
	//返回好友信息，组装数据
	userRecords := make(map[string]*userclient.UserEntity, len(users.User))
	for i, _ := range users.User {
		userRecords[users.User[i].Id] = users.User[i]
	}
	respList := make([]*types.Friends, 0, len(friends.List))
	for _, v := range friends.List {
		friend := &types.Friends{
			Id:        v.Id,
			FriendUid: v.FriendUid,
		}
		if u, ok := userRecords[v.FriendUid]; ok {
			friend.Nickname = u.Nickname
			friend.Avatar = u.Avatar

		}
		respList = append(respList, friend)
	}

	return &types.FriendListResp{
		List: respList,
	}, nil
}
