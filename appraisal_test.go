package evepraisal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newTestAppraisal(created time.Time, expireTime time.Time, expireMinutes int64) *Appraisal {
	return &Appraisal{
		ID:            "test",
		Created:       created.Unix(),
		ExpireTime:    &expireTime,
		ExpireMinutes: expireMinutes,
	}
}

func TestAppraisalExpiration(t *testing.T) {
	now := time.Now()

	// Not expired
	assert.False(t, newTestAppraisal(time.Now(), time.Now().Add(time.Second), 1).IsExpired(now, now.Add(-59*time.Second)))

	// Expired because of expire time
	assert.True(t, newTestAppraisal(time.Now(), time.Now().Add(-2*time.Second), 100).IsExpired(now, now))

	// Expired because of expire minutes
	assert.True(t, newTestAppraisal(time.Now(), time.Now().Add(time.Second), 1).IsExpired(now, now.Add(-2*time.Minute)))
}
