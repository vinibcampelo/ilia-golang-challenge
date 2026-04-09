package transaction

import "errors"

type Type string

const (
	TypeCredit Type = "CREDIT"
	TypeDebit  Type = "DEBIT"
)

var ErrInvalidType = errors.New("invalid transaction type")

func ParseType(s string) (Type, error) {
	switch Type(s) {
	case TypeCredit, TypeDebit:
		return Type(s), nil
	default:
		return "", ErrInvalidType
	}
}

func (t Type) String() string {
	return string(t)
}

func (t Type) IsDebit() bool {
	return t == TypeDebit
}
