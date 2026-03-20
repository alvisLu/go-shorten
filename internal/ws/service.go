package ws

type WsReq struct {
	Type string `json:"type"`
}

type WsResp struct {
	Status string `json:"status"`
}

type WsService interface {
	WsHealth() WsResp
}

type wsService struct{}

func NewService() WsService {
	return &wsService{}
}

func (s *wsService) WsHealth() WsResp {
	return WsResp{Status: "ok"}
}
