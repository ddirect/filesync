package records

import (
	"time"
)

func (cm *CacheMeta) Time() time.Time {
	return time.Unix(0, cm.GetTimeNs())
}

func NewCacheMeta(path string, device uint64) *CacheMeta {
	return &CacheMeta{Path: path, TimeNs: time.Now().UnixNano(), Device: device}
}
