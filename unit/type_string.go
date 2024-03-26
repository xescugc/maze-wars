// Code generated by "enumer -type=Type -transform=lower -transform=snake -output=type_string.go"; DO NOT EDIT.

package unit

import (
	"fmt"
	"strings"
)

const _TypeName = "spiritflamoctopusracconcyclopeeyebeastbutterflymoleskullsnake"

var _TypeIndex = [...]uint8{0, 6, 10, 17, 23, 30, 33, 38, 47, 51, 56, 61}

const _TypeLowerName = "spiritflamoctopusracconcyclopeeyebeastbutterflymoleskullsnake"

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
	_ = x[Spirit-(0)]
	_ = x[Flam-(1)]
	_ = x[Octopus-(2)]
	_ = x[Raccon-(3)]
	_ = x[Cyclope-(4)]
	_ = x[Eye-(5)]
	_ = x[Beast-(6)]
	_ = x[Butterfly-(7)]
	_ = x[Mole-(8)]
	_ = x[Skull-(9)]
	_ = x[Snake-(10)]
}

var _TypeValues = []Type{Spirit, Flam, Octopus, Raccon, Cyclope, Eye, Beast, Butterfly, Mole, Skull, Snake}

var _TypeNameToValueMap = map[string]Type{
	_TypeName[0:6]:        Spirit,
	_TypeLowerName[0:6]:   Spirit,
	_TypeName[6:10]:       Flam,
	_TypeLowerName[6:10]:  Flam,
	_TypeName[10:17]:      Octopus,
	_TypeLowerName[10:17]: Octopus,
	_TypeName[17:23]:      Raccon,
	_TypeLowerName[17:23]: Raccon,
	_TypeName[23:30]:      Cyclope,
	_TypeLowerName[23:30]: Cyclope,
	_TypeName[30:33]:      Eye,
	_TypeLowerName[30:33]: Eye,
	_TypeName[33:38]:      Beast,
	_TypeLowerName[33:38]: Beast,
	_TypeName[38:47]:      Butterfly,
	_TypeLowerName[38:47]: Butterfly,
	_TypeName[47:51]:      Mole,
	_TypeLowerName[47:51]: Mole,
	_TypeName[51:56]:      Skull,
	_TypeLowerName[51:56]: Skull,
	_TypeName[56:61]:      Snake,
	_TypeLowerName[56:61]: Snake,
}

var _TypeNames = []string{
	_TypeName[0:6],
	_TypeName[6:10],
	_TypeName[10:17],
	_TypeName[17:23],
	_TypeName[23:30],
	_TypeName[30:33],
	_TypeName[33:38],
	_TypeName[38:47],
	_TypeName[47:51],
	_TypeName[51:56],
	_TypeName[56:61],
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
