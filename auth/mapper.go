package auth

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/storage"
)

type Mapper interface {
	ToOutPutAccount(account *storage.Account) *OutputAccount
	ToOutPutAccounts(arr []*storage.Account) []*OutputAccount
}

type mapper struct {
}

func newMapper() Mapper {

	return &mapper{}
}
func (o *mapper) ToOutPutAccount(m *storage.Account) *OutputAccount {
	if m == nil {
		return nil
	}
	addr, _ := address.NewFromString(m.Miner)
	return &OutputAccount{
		Id:         m.Id,
		Name:       m.Name,
		Miner:      addr,
		Comment:    m.Comment,
		State:      m.State,
		SourceType: m.SourceType,
		CreateTime: m.CreateTime.Unix(),
		UpdateTime: m.UpdateTime.Unix(),
		ReqLimit:   m.ReqLimit}
}

func (o *mapper) ToOutPutAccounts(arr []*storage.Account) []*OutputAccount {
	list := make([]*OutputAccount, 0, len(arr))
	for _, v := range arr {
		list = append(list, o.ToOutPutAccount(v))
	}
	return list
}
