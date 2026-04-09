package transaction

import "errors"

type MinorAmount struct {
	minorUnits int64
}

var ErrInvalidMinorAmount = errors.New("amount must be positive")

func NewMinorAmount(minorUnits int64) (MinorAmount, error) {
	if minorUnits <= 0 {
		return MinorAmount{}, ErrInvalidMinorAmount
	}
	return MinorAmount{minorUnits: minorUnits}, nil
}

func RehydrateMinorAmount(minorUnits int64) MinorAmount {
	return MinorAmount{minorUnits: minorUnits}
}

func (m MinorAmount) MinorUnits() int64 {
	return m.minorUnits
}

func (m MinorAmount) CanCoverDebit(balanceMinorUnits int64) bool {
	return balanceMinorUnits >= m.minorUnits
}
