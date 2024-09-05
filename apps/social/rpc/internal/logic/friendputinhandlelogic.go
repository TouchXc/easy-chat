package logic

import (
	"context"
	"database/sql"
	"easy-chat/apps/social/rpc/internal/svc"
	"easy-chat/apps/social/rpc/social"
	"easy-chat/apps/social/socialmodels"
	"easy-chat/pkg/constants"
	"easy-chat/pkg/xerr"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrFriendReqBeforePass   = xerr.NewMsg("好友申请已通过")
	ErrFriendReqBeforeRefuse = xerr.NewMsg("好友申请已被拒绝")
)

type FriendPutInHandleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFriendPutInHandleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FriendPutInHandleLogic {
	return &FriendPutInHandleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FriendPutInHandleLogic) FriendPutInHandle(in *social.FriendPutInHandleReq) (*social.FriendPutInHandleResp, error) {
	//获取好友申请记录
	friendReq, err := l.svcCtx.FriendRequestsModel.FindOne(l.ctx, int64(in.FriendReqId))
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "find friendsRequest by FriendReqId err: %v req: %v", err, in)
	}
	//验证是否有处理
	switch constants.HandlerResult(friendReq.HandleResult.Int64) {
	case constants.PassHandlerResult:
		return nil, errors.WithStack(ErrFriendReqBeforePass)
	case constants.RefuseHandlerResult:
		return nil, errors.WithStack(ErrFriendReqBeforeRefuse)
	}
	friendReq.HandleResult.Int64 = int64(in.HandleResult)
	//修改申请结果 -》通过【建立两条好友关系记录表】-》事务
	err = l.svcCtx.FriendRequestsModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err = l.svcCtx.FriendRequestsModel.Update(l.ctx, session, friendReq); err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "update friend request err: %v req: %v", err, in)
		}
		if constants.HandlerResult(in.HandleResult) != constants.PassHandlerResult {
			return nil
		}
		friends := []*socialmodels.Friends{
			{
				UserId:    friendReq.UserId,
				FriendUid: friendReq.ReqUid,
				Remark:    sql.NullString{Valid: false},
				AddSource: sql.NullInt64{Valid: false},
				CreatedAt: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			}, {
				UserId:    friendReq.ReqUid,
				FriendUid: friendReq.UserId,
				Remark:    sql.NullString{Valid: false},
				AddSource: sql.NullInt64{Valid: false},
				CreatedAt: sql.NullTime{
					Time:  time.Now(),
					Valid: true,
				},
			},
		}
		// 插入好友关系记录
		_, err = l.svcCtx.FriendsModel.Inserts(l.ctx, session, friends...)
		fmt.Println(err)
		if err != nil {
			return errors.Wrapf(xerr.NewDBErr(), "friends inserts err: %v req: %v", err, in)
		}
		return nil
	})
	return &social.FriendPutInHandleResp{}, nil
}
