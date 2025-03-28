// Code generated by "enumer -type=Type -transform=lower -transform=snake -output=type_string.go"; DO NOT EDIT.

package unit

import (
	"fmt"
	"strings"
)

const _TypeName = "ninjastatuehunterslimemoleskeleton_demonbutterflyblend_masterrobotmonkey_boxer"

var _TypeIndex = [...]uint8{0, 5, 11, 17, 22, 26, 40, 49, 61, 66, 78}

const _TypeLowerName = "ninjastatuehunterslimemoleskeleton_demonbutterflyblend_masterrobotmonkey_boxer"

func (i Type) String() string {
	if i < 0 || i >= Type(len(_TypeIndex)-1) {
		return fmt.Sprintf("Type(%d)", i)
	}
	return _TypeName[_TypeIndex[i]:_TypeIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _TypeNoOp() {
	var x [1]struct{}
	_ = x[Ninja-(0)]
	_ = x[Statue-(1)]
	_ = x[Hunter-(2)]
	_ = x[Slime-(3)]
	_ = x[Mole-(4)]
	_ = x[SkeletonDemon-(5)]
	_ = x[Butterfly-(6)]
	_ = x[BlendMaster-(7)]
	_ = x[Robot-(8)]
	_ = x[MonkeyBoxer-(9)]
}

var _TypeValues = []Type{Ninja, Statue, Hunter, Slime, Mole, SkeletonDemon, Butterfly, BlendMaster, Robot, MonkeyBoxer}

var _TypeNameToValueMap = map[string]Type{
	_TypeName[0:5]:        Ninja,
	_TypeLowerName[0:5]:   Ninja,
	_TypeName[5:11]:       Statue,
	_TypeLowerName[5:11]:  Statue,
	_TypeName[11:17]:      Hunter,
	_TypeLowerName[11:17]: Hunter,
	_TypeName[17:22]:      Slime,
	_TypeLowerName[17:22]: Slime,
	_TypeName[22:26]:      Mole,
	_TypeLowerName[22:26]: Mole,
	_TypeName[26:40]:      SkeletonDemon,
	_TypeLowerName[26:40]: SkeletonDemon,
	_TypeName[40:49]:      Butterfly,
	_TypeLowerName[40:49]: Butterfly,
	_TypeName[49:61]:      BlendMaster,
	_TypeLowerName[49:61]: BlendMaster,
	_TypeName[61:66]:      Robot,
	_TypeLowerName[61:66]: Robot,
	_TypeName[66:78]:      MonkeyBoxer,
	_TypeLowerName[66:78]: MonkeyBoxer,
}

var _TypeNames = []string{
	_TypeName[0:5],
	_TypeName[5:11],
	_TypeName[11:17],
	_TypeName[17:22],
	_TypeName[22:26],
	_TypeName[26:40],
	_TypeName[40:49],
	_TypeName[49:61],
	_TypeName[61:66],
	_TypeName[66:78],
}

// TypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func TypeString(s string) (Type, error) {
	if val, ok := _TypeNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _TypeNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Type values", s)
}

// TypeValues returns all values of the enum
func TypeValues() []Type {
	return _TypeValues
}

// TypeStrings returns a slice of all String values of the enum
func TypeStrings() []string {
	strs := make([]string, len(_TypeNames))
	copy(strs, _TypeNames)
	return strs
}

// IsAType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Type) IsAType() bool {
	for _, v := range _TypeValues {
		if i == v {
			return true
		}
	}
	return false
}
