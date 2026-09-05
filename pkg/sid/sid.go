// Package sid 提供基于 sonyflake 的分布式唯一 ID 生成。
package sid

import (
	"errors"

	"github.com/samber/do/v2"
	"github.com/sony/sonyflake"
)

// Sid 封装 sonyflake 分布式 ID 生成器。
type Sid struct {
	sf *sonyflake.Sonyflake
}

// Package registers the sid provider into the injector.
var Package = do.Package(do.Lazy(New))

// New 构造 ID 生成器，由注入容器调用。
func New(i do.Injector) (*Sid, error) {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	if sf == nil {
		return nil, errors.New("sonyflake not created")
	}
	return &Sid{sf}, nil
}

// GenString 生成 base62 编码的字符串 ID。
func (s Sid) GenString() (string, error) {
	id, err := s.sf.NextID()
	if err != nil {
		return "", err
	}
	return IntToBase62(int(id)), nil
}

// GenUint64 生成原始 uint64 ID。
func (s Sid) GenUint64() (uint64, error) {
	return s.sf.NextID()
}
