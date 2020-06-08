package staticdump

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPatch_18_04RigOverride(t *testing.T) {
	{
		alias, found := patch_18_04RigOverride("Small Anti-EM Screen Reinforcer I")
		assert.Equal(t, "Small EM Shield Reinforcer I", alias)
		assert.True(t, found)
	}
}

func TestComputeAliases(t *testing.T) {
	{
		override, aliases := computeAliases(0, "Medium Anti-EM Screen Reinforcer I")
		assert.Equal(t, "Medium EM Shield Reinforcer I", override)
		assert.Equal(t, []string{"Medium Anti-EM Screen Reinforcer I"}, aliases)
	}
}
